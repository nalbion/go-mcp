package mcp_test

import (
	"context"
	"testing"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp"
	mcpserver "github.com/nalbion/go-mcp/pkg/mcp/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to convert string to pointer
func strPtr(s string) *string {
	return &s
}

// TestToolRegistration tests the registration and management of tools
func TestToolRegistration(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := mcpserver.NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverInfo, &options)

	// when we add a tool
	err := mcpServer.AddTool(
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
		func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
			return mcp.CallToolResult{}, nil
		},
	)

	// then the tool is added successfully
	require.NoError(t, err)

	// when we try to add a tool with the same name
	err = mcpServer.AddTool(
		"test-tool",
		"Another test tool",
		mcp.ToolInputSchema{},
		func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
			return mcp.CallToolResult{}, nil
		},
	)

	// then we get an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestToolExecution tests the execution of tools
func TestToolExecution(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := mcpserver.NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverInfo, &options)

	// Add a test tool that echoes back the input
	err := mcpServer.AddTool(
		"echo",
		"Echoes back the input",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]mcp.ToolInputSchemaProperty{
				"message": {
					Type:        "string",
					Description: "Message to echo",
				},
			},
		},
		func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
			message, _ := params.Arguments["message"].(string)
			return mcp.CallToolResult{
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": message,
					},
				},
			}, nil
		},
	)
	require.NoError(t, err)

	// when we call the tool
	result, err := mcpServer.HandleCallTool(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "tools/call",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.CallToolRequestParams{
				Name: "echo",
				Arguments: map[string]interface{}{
					"message": "Hello, world!",
				},
			},
		},
	}, nil)

	// then the tool executes successfully
	require.NoError(t, err)
	callResult, ok := result.AdditionalProperties.(mcp.CallToolResult)
	require.True(t, ok)
	require.Len(t, callResult.Content, 1)
	content, ok := callResult.Content[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "text", content["type"])
	assert.Equal(t, "Hello, world!", content["text"])
}

// TestToolNotification tests the tool list changed notification
func TestToolNotification(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := mcpserver.NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverInfo, &options)

	// Create a mock transport to capture notifications
	mockTransport := &mockTransport{
		notificationCh: make(chan *jsonrpc.JSONRPCNotification, 1),
	}
	mcpServer.Connect(ctx, mockTransport)

	// when we send a tool list changed notification
	err := mcpServer.SendToolListChanged()

	// then the notification is sent successfully
	require.NoError(t, err)

	// and the notification is received
	select {
	case notification := <-mockTransport.notificationCh:
		assert.Equal(t, "tools/listChanged", notification.Method)
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for notification")
	}
}

// TestToolValidation tests the validation of tool input
func TestToolValidation(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := mcpserver.NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverInfo, &options)

	// Add a test tool with required parameters
	err := mcpServer.AddTool(
		"validate",
		"A tool with required parameters",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]mcp.ToolInputSchemaProperty{
				"required_param": {
					Type:        "string",
					Description: "A required parameter",
				},
			},
			Required: []string{"required_param"},
		},
		func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
			return mcp.CallToolResult{}, nil
		},
	)
	require.NoError(t, err)

	// when we call the tool without the required parameter
	result, err := mcpServer.HandleCallTool(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "tools/call",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.CallToolRequestParams{
				Name:      "validate",
				Arguments: map[string]interface{}{},
			},
		},
	}, nil)

	// then we get an error
	assert.Error(t, err)
	assert.Empty(t, result.AdditionalProperties)

	// when we call the tool with the required parameter
	result, err = mcpServer.HandleCallTool(ctx, &jsonrpc.JSONRPCRequest{
		Id:     2,
		Method: "tools/call",
		Params: &jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: mcp.CallToolRequestParams{
				Name: "validate",
				Arguments: map[string]interface{}{
					"required_param": "value",
				},
			},
		},
	}, nil)

	// then the tool executes successfully
	require.NoError(t, err)
	_, ok := result.AdditionalProperties.(mcp.CallToolResult)
	require.True(t, ok)
}

// Mock transport for testing notifications
type mockTransport struct {
	notificationCh chan *jsonrpc.JSONRPCNotification
}

func (m *mockTransport) Start() error {
	return nil
}

func (m *mockTransport) Close() error {
	return nil
}

func (m *mockTransport) Send(message jsonrpc.JSONRPCMessage) error {
	return nil
}

func (m *mockTransport) SetOnClose(func()) {
}

func (m *mockTransport) SetOnError(func(err error)) {
}

func (m *mockTransport) SetOnMessage(func(message jsonrpc.JSONRPCMessage)) {
}

// TestMultipleTools tests adding and listing multiple tools
func TestMultipleTools(t *testing.T) {
	// given
	ctx := context.Background()
	serverInfo := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	options := mcpserver.NewServerOptions()
	options.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverInfo, &options)

	// Add multiple tools
	tools := []mcpserver.RegisteredTool{
		{
			Tool: mcp.Tool{
				Name:        "tool1",
				Description: "Tool 1",
				InputSchema: mcp.ToolInputSchema{},
			},
			Handler: func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
				return mcp.CallToolResult{}, nil
			},
		},
		{
			Tool: mcp.Tool{
				Name:        "tool2",
				Description: "Tool 2",
				InputSchema: mcp.ToolInputSchema{},
			},
			Handler: func(params mcp.CallToolRequestParams) (mcp.CallToolResult, error) {
				return mcp.CallToolResult{}, nil
			},
		},
	}

	// when we add multiple tools
	err := mcpServer.AddTools(tools)

	// then the tools are added successfully
	require.NoError(t, err)

	// when we list the tools
	result, err := mcpServer.HandleListTools(ctx, &jsonrpc.JSONRPCRequest{
		Id:     1,
		Method: "tools/list",
		Params: &jsonrpc.JSONRPCRequestParams{},
	}, nil)

	// then we get all the tools
	require.NoError(t, err)
	listResult, ok := result.AdditionalProperties.(mcp.ListToolsResult)
	require.True(t, ok)
	assert.Len(t, listResult.Tools, 2)

	// Verify tool names
	toolNames := []string{listResult.Tools[0].Name, listResult.Tools[1].Name}
	assert.Contains(t, toolNames, "tool1")
	assert.Contains(t, toolNames, "tool2")
}
