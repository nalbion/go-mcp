package server

import (
	"context"
	"testing"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock tool handler
type mockToolHandler struct {
	called bool
	params mcp.CallToolRequestParams
}

func (m *mockToolHandler) Handle(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
	m.called = true
	m.params = params
	return mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "Mock tool response",
			},
		},
	}, nil
}

// Mock prompt provider
type mockPromptProvider struct {
	called bool
	params mcp.GetPromptRequestParams
}

func (m *mockPromptProvider) Provide(params mcp.GetPromptRequestParams) mcp.GetPromptResult {
	m.called = true
	m.params = params
	return mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleAssistant,
				Content: mcp.TextContent{
					Type: "text",
					Text: "You are a test assistant",
				},
			},
		},
	}
}

// Mock resource handler
type mockResourceHandler struct {
	called bool
	params mcp.ReadResourceRequestParams
}

func (m *mockResourceHandler) Read(params mcp.ReadResourceRequestParams) mcp.ReadResourceResult {
	m.called = true
	m.params = params
	return mcp.ReadResourceResult{
		Contents: []mcp.HasResourceContents{
			mcp.NewTextResourceContents("Test resource content", "text/plain"),
		},
	}
}

func TestServerInitialization(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}

	// when
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools:     &mcp.ServerToolsCapabilities{},
		Prompts:   &mcp.ServerCapabilitiesPrompts{},
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	// then
	assert.NotNil(t, server)
	assert.Equal(t, serverInfo, server.serverInfo)
	assert.Equal(t, options.Capabilities, server.capabilities)
}

func TestHandleInitialize(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools:     &mcp.ServerToolsCapabilities{},
		Prompts:   &mcp.ServerCapabilitiesPrompts{},
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	initParams := mcp.InitializeRequestParams{
		ProtocolVersion: "1.0",
		ClientInfo: mcp.Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		},
		Capabilities: mcp.ClientCapabilities{
			Experimental: mcp.ClientCapabilitiesExperimental{},
			Roots:        &mcp.ClientCapabilitiesRoots{},
			Sampling:     mcp.ClientCapabilitiesSampling{},
		},
	}

	// when
	result, err := server.handleInitialize(ctx, &jsonrpc.JSONRPCRequest{
		Id:     jsonrpc.RequestId(1),
		Method: "initialize",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: initParams,
		},
	}, nil)

	// then
	require.NoError(t, err)
	initResult, ok := result.AdditionalProperties.(mcp.InitializeResult)
	require.True(t, ok)
	assert.Equal(t, "1.0", initResult.ProtocolVersion)
	assert.Equal(t, serverInfo, initResult.ServerInfo)
	assert.Equal(t, options.Capabilities, initResult.Capabilities)
}

func TestAddTool(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	server := NewServer(ctx, serverInfo, &options)

	toolHandler := &mockToolHandler{}

	// when
	err := server.AddTool(
		"test-tool",
		"A test tool",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]mcp.ToolInputSchemaProperty{
				"param1": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
		},
		toolHandler.Handle,
	)

	// then
	require.NoError(t, err)
	server.toolsMutex.RLock()
	tool, exists := server.tools["test-tool"]
	server.toolsMutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, "test-tool", tool.Tool.Name)
	assert.Equal(t, "A test tool", tool.Tool.Description)
}

func TestHandleListTools(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	server := NewServer(ctx, serverInfo, &options)

	toolHandler := &mockToolHandler{}
	err := server.AddTool(
		"test-tool",
		"A test tool",
		mcp.ToolInputSchema{},
		toolHandler.Handle,
	)
	require.NoError(t, err)

	// when
	result, err := server.HandleListTools(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "tools/list",
	}, nil)

	// then
	require.NoError(t, err)
	listResult, ok := result.AdditionalProperties.(mcp.ListToolsResult)
	require.True(t, ok)
	assert.Len(t, listResult.Tools, 1)
	assert.Equal(t, "test-tool", listResult.Tools[0].Name)
}

func TestHandleCallTool(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	server := NewServer(ctx, serverInfo, &options)

	toolHandler := &mockToolHandler{}
	err := server.AddTool(
		"test-tool",
		"A test tool",
		mcp.ToolInputSchema{},
		toolHandler.Handle,
	)
	require.NoError(t, err)

	// when
	result, err := server.HandleCallTool(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "tools/call",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.CallToolRequestParams{
				Name: "test-tool",
				Arguments: map[string]interface{}{
					"param1": "value1",
				},
			},
		},
	}, nil)

	// then
	require.NoError(t, err)
	callResult, ok := result.AdditionalProperties.(mcp.CallToolResult)
	require.True(t, ok)
	assert.Len(t, callResult.Content, 1)
	assert.True(t, toolHandler.called)
	assert.Equal(t, "test-tool", toolHandler.params.Name)
	assert.Equal(t, "value1", toolHandler.params.Arguments["param1"])
}

func TestAddPrompt(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Prompts: &mcp.ServerCapabilitiesPrompts{},
	}
	server := NewServer(ctx, serverInfo, &options)

	promptProvider := &mockPromptProvider{}

	// when
	err := server.AddPrompt(
		mcp.Prompt{
			Name:        "Test Prompt",
			Description: "A test prompt",
		},
		promptProvider.Provide,
	)

	// then
	require.NoError(t, err)
	server.promptsMutex.RLock()
	prompt, exists := server.prompts["test-prompt"]
	server.promptsMutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, "Test Prompt", prompt.Prompt.Name)
}

func TestHandleListPrompts(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Prompts: &mcp.ServerCapabilitiesPrompts{},
	}
	server := NewServer(ctx, serverInfo, &options)

	promptProvider := &mockPromptProvider{}
	err := server.AddPrompt(
		mcp.Prompt{
			Name:        "Test Prompt",
			Description: "A test prompt",
		},
		promptProvider.Provide,
	)
	require.NoError(t, err)

	// when
	result, err := server.handleListPrompts(ctx, &jsonrpc.JSONRPCRequest{
		Method: "prompts/list",
	}, nil)

	// then
	require.NoError(t, err)
	listResult, ok := result.AdditionalProperties.(mcp.ListPromptsResult)
	require.True(t, ok)
	assert.Len(t, listResult.Prompts, 1)
}

func TestHandleGetPrompt(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Prompts: &mcp.ServerCapabilitiesPrompts{},
	}
	server := NewServer(ctx, serverInfo, &options)

	promptProvider := &mockPromptProvider{}
	err := server.AddPrompt(
		mcp.Prompt{
			Name:        "Test Prompt",
			Description: "A test prompt",
		},
		promptProvider.Provide,
	)
	require.NoError(t, err)

	// when
	result, err := server.handleGetPrompt(ctx, &jsonrpc.JSONRPCRequest{
		Method: "prompts/get",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.GetPromptRequestParams{
				Name: "test-prompt",
			},
		},
	}, nil)

	// then
	require.NoError(t, err)
	getResult, ok := result.AdditionalProperties.(mcp.GetPromptResult)
	require.True(t, ok)
	assert.Equal(t, "Test Prompt", getResult.Messages[0].Content)
	assert.True(t, promptProvider.called)
	assert.Equal(t, "test-prompt", promptProvider.params.Name)
}

func TestAddResource(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	resourceHandler := &mockResourceHandler{}
	description := "A test resource"
	mimeType := "text/plain"

	// when
	err := server.AddResource(
		"test://resource",
		"Test Resource",
		description,
		mimeType,
		resourceHandler.Read,
	)

	// then
	require.NoError(t, err)
	server.resourcesMutex.RLock()
	resource, exists := server.resources["test://resource"]
	server.resourcesMutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, "test://resource", resource.Resource.Uri)
	assert.Equal(t, "Test Resource", resource.Resource.Name)
}

func TestHandleListResources(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	resourceHandler := &mockResourceHandler{}
	description := "A test resource"
	mimeType := "text/plain"
	err := server.AddResource(
		"test://resource",
		"Test Resource",
		description,
		mimeType,
		resourceHandler.Read,
	)
	require.NoError(t, err)

	// when
	result, err := server.handleListResources(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "resources/list",
	}, nil)

	// then
	require.NoError(t, err)
	listResult, ok := result.AdditionalProperties.(mcp.ListResourcesResult)
	require.True(t, ok)
	assert.Len(t, listResult.Resources, 1)
	assert.Equal(t, "test://resource", listResult.Resources[0].Uri)
}

func TestHandleReadResource(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	resourceHandler := &mockResourceHandler{}
	description := "A test resource"
	mimeType := "text/plain"
	err := server.AddResource(
		"test://resource",
		"Test Resource",
		description,
		mimeType,
		resourceHandler.Read,
	)
	require.NoError(t, err)

	// when
	result, err := server.handleReadResource(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "resources/read",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.ReadResourceRequestParams{
				Uri: "test://resource",
			},
		},
	}, nil)

	// then
	require.NoError(t, err)
	readResult, ok := result.AdditionalProperties.(mcp.ReadResourceResult)
	require.True(t, ok)
	assert.Equal(t, "Test resource content", readResult.Contents[0].GetValue())
	assert.Equal(t, "text/plain", readResult.Contents[0].GetMimeType())
	assert.True(t, resourceHandler.called)
	assert.Equal(t, "test://resource", resourceHandler.params.Uri)
}

func TestSendNotifications(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools:     &mcp.ServerCapabilitiesTools{},
		Prompts:   &mcp.ServerCapabilitiesPrompts{},
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	server := NewServer(ctx, serverInfo, &options)

	mockTransport := &jsonrpc.MockTransport{}
	server.Connect(ctx, mockTransport)

	// when/then for each notification type
	t.Run("SendToolListChanged", func(t *testing.T) {
		// when
		err := server.SendToolListChanged()

		// then
		require.NoError(t, err)
		assert.Len(t, mockTransport.SentNotifications, 1)
		assert.Equal(t, "tools/listChanged", mockTransport.SentNotifications[0].Method)
	})

	// Reset notifications for next test
	mockTransport.SentNotifications = nil

	t.Run("SendPromptListChanged", func(t *testing.T) {
		// when
		err := server.SendPromptListChanged()

		// then
		require.NoError(t, err)
		assert.Len(t, mockTransport.SentNotifications, 1)
		assert.Equal(t, "prompts/listChanged", mockTransport.SentNotifications[0].Method)
	})

	// Reset notifications for next test
	mockTransport.SentNotifications = nil

	t.Run("SendResourceListChanged", func(t *testing.T) {
		// when
		err := server.SendResourceListChanged()

		// then
		require.NoError(t, err)
		assert.Len(t, mockTransport.SentNotifications, 1)
		assert.Equal(t, "resources/listChanged", mockTransport.SentNotifications[0].Method)
	})

	// Reset notifications for next test
	mockTransport.SentNotifications = nil

	t.Run("SendResourceUpdated", func(t *testing.T) {
		// when
		err := server.SendResourceUpdated(mcp.ResourceUpdatedNotificationParams{
			Uri: "test://resource",
		})

		// then
		require.NoError(t, err)
		assert.Len(t, mockTransport.SentNotifications, 1)
		assert.Equal(t, "resources/updated", mockTransport.SentNotifications[0].Method)
	})
}
