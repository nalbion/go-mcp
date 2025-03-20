package mcp_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc/client"
	"github.com/nalbion/go-mcp/pkg/jsonrpc/server"
	"github.com/nalbion/go-mcp/pkg/mcp"
	mcpclient "github.com/nalbion/go-mcp/pkg/mcp/client"
	mcpserver "github.com/nalbion/go-mcp/pkg/mcp/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration tests the full stack with client and server communicating over stdio transport
func TestIntegration(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create server-side transport
	serverTransport := server.NewStdioServerTransport(ctx, os.Stdin, os.Stdout)

	// Create server
	serverImpl := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	serverOptions := mcpserver.NewServerOptions()
	serverOptions.Capabilities = mcp.ServerCapabilities{
		Tools: &mcp.ServerToolsCapabilities{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverImpl, &serverOptions)

	// Add a test tool to the server
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

	// Connect server to transport
	// serverTransport.SetRequestHandler = mcpServer.SetRequestHandler
	// serverTransport.SetNotificationHandler = mcpServer.SetNotificationHandler

	// Start server
	err = serverTransport.Start()
	require.NoError(t, err)
	defer serverTransport.Close()

	// Create client-side transport
	clientTransport := client.NewStdioClientTransport(ctx, client.StdioServerParameters{})

	// Create client
	clientImpl := mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}
	mcpClient := mcpclient.NewClient(ctx, clientImpl, mcpclient.ClientOptions{
		Capabilities: mcp.ClientCapabilities{},
	})

	// when we connect the client to the server
	err = mcpClient.Connect(clientTransport)

	// then the connection is successful
	require.NoError(t, err)

	// Test listing tools
	t.Run("ListTools", func(t *testing.T) {
		// when we list the tools
		result := &mcp.ListToolsResult{}
		err = mcpClient.ListTools(mcp.ListToolsRequestParams{}, result, nil)

		// then we get the tools successfully
		require.NoError(t, err)
		require.Len(t, result.Tools, 1)
		assert.Equal(t, "echo", result.Tools[0].Name)
	})

	// Test calling a tool
	t.Run("CallTool", func(t *testing.T) {
		// given a message to echo
		message := "Hello, MCP!"

		// when we call the echo tool
		result := &mcp.CallToolResult{}
		err = mcpClient.CallTool(mcp.CallToolRequestParams{
			Name: "echo",
			Arguments: map[string]interface{}{
				"message": message,
			},
		}, result, nil)

		// then we get the expected response
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		content, ok := result.Content[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "text", content["type"])
		assert.Equal(t, message, content["text"])
	})
}

// TestClientServerCapabilityNegotiation tests that the client and server correctly negotiate capabilities
func TestClientServerCapabilityNegotiation(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create server with full capabilities
	serverTransport := server.NewStdioServerTransport(ctx, os.Stdin, os.Stdout)
	serverImpl := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	serverOptions := mcpserver.NewServerOptions()
	serverOptions.Capabilities = mcp.ServerCapabilities{
		Tools:     &mcp.ServerToolsCapabilities{},
		Prompts:   &mcp.ServerCapabilitiesPrompts{},
		Resources: &mcp.ServerCapabilitiesResources{},
	}
	mcpServer := mcpserver.NewServer(ctx, serverImpl, &serverOptions)

	// Connect server to transport
	mcpServer.Connect(ctx, serverTransport)

	// Start server
	err := serverTransport.Start()
	require.NoError(t, err)
	defer serverTransport.Close()

	// Create client with limited capabilities
	clientTransport := client.NewStdioClientTransport(ctx, client.StdioServerParameters{})
	clientImpl := mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}

	// when we create a client
	mcpClient := mcpclient.NewClient(ctx, clientImpl, mcpclient.ClientOptions{
		Capabilities: mcp.ClientCapabilities{},
	})

	// then the connection is successful
	err = mcpClient.Connect(clientTransport)
	require.NoError(t, err)

	// Test that we can use the tools capability
	t.Run("ToolsCapability", func(t *testing.T) {
		// when we try to list tools
		result := &mcp.ListToolsResult{}
		err = mcpClient.ListTools(mcp.ListToolsRequestParams{}, result, nil)

		// then it succeeds
		require.NoError(t, err)
	})

	// Test that we cannot use the prompts capability
	t.Run("PromptsCapability", func(t *testing.T) {
		// when we try to list prompts
		result := &mcp.ListPromptsResult{}
		err = mcpClient.ListPrompts(mcp.ListPromptsRequestParams{}, result, nil)

		// then it fails because we didn't declare the capability
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})
}

// TestPing tests the ping functionality
func TestPing(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create server
	serverTransport := server.NewStdioServerTransport(ctx, os.Stdin, os.Stdout)
	serverImpl := mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	serverOptions := mcpserver.NewServerOptions()
	mcpServer := mcpserver.NewServer(ctx, serverImpl, &serverOptions)

	// Connect server to transport
	mcpServer.Connect(ctx, serverTransport)

	// Start server
	err := serverTransport.Start()
	require.NoError(t, err)
	defer serverTransport.Close()

	// Create client
	clientTransport := client.NewStdioClientTransport(ctx, client.StdioServerParameters{})
	clientImpl := mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}
	mcpClient := mcpclient.NewClient(ctx, clientImpl, mcpclient.ClientOptions{})

	// Connect client to server
	err = mcpClient.Connect(clientTransport)
	require.NoError(t, err)

	// when we ping the server
	err = mcpClient.Ping(&mcp.RequestOptions{})

	// then the ping succeeds
	require.NoError(t, err)
}
