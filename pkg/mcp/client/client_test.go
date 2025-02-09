package client

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	jsonrpc_client "github.com/nalbion/go-mcp/pkg/jsonrpc/client"
	"github.com/nalbion/go-mcp/pkg/mcp"
	"github.com/stretchr/testify/require"
)

func TestIntegrationClient(t *testing.T) {
	// given
	ctx := context.Background()
	client := NewClient(ctx,
		mcp.Implementation{
			Name:    "go-mcp",
			Version: "0.0.1",
		},
		ClientOptions{
			Capabilities: mcp.ClientCapabilities{},
		},
	)

	goMcpRootPath, err := filepath.Abs("../../..")
	require.NoError(t, err)
	transport := jsonrpc_client.NewStdioClientTransport(ctx,
		jsonrpc_client.StdioServerParameters{
			Command: "docker",
			Args: []string{
				"run",
				"-i",
				"--rm",
				"--mount", fmt.Sprintf("type=bind,src=%s,dst=/projects/go-mcp", goMcpRootPath),
				"mcp/filesystem",
				"/projects",
			},
		},
	)

	// when we connect to the filesystem server
	err = client.Connect(transport)
	require.NoError(t, err)

	t.Run("ListTools", func(t *testing.T) {
		// then we can list the tools provided by the server
		result := &mcp.ListToolsResult{}
		err = client.ListTools(mcp.ListToolsRequestParams{}, result, nil)
		require.NoError(t, err)
		require.NotEmpty(t, result.Tools)
	})

	t.Run("directory_tree", func(t *testing.T) {
		// when we call the directory_tree tool
		result := &mcp.CallToolResult{}
		err = client.CallTool(
			mcp.CallToolRequestParams{
				Name: "directory_tree",
				Arguments: mcp.CallToolRequestParamsArguments{
					"path": "/projects",
				},
			}, result, nil)

		// then
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		if content, ok := result.Content[0].(map[string]any); !ok {
			t.Fatalf("unexpected content type: %T", result.Content[0])
		} else {
			text := content["text"]
			require.NotEmpty(t, text)
			// fmt.Println(text)
		}
	})
}
