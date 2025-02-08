package client

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	jsonrpc_client "github.com/nalbion/go-mcp/pkg/jsonrpc/client"
	"github.com/nalbion/go-mcp/pkg/mcp/shared"
	"github.com/stretchr/testify/require"
)

func TestIntegrationClient(t *testing.T) {
	// given
	ctx := context.Background()
	client := NewClient(ctx,
		shared.Implementation{
			Name:    "go-mcp",
			Version: "0.0.1",
		},
		ClientOptions{
			Capabilities: shared.ClientCapabilities{},
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

	// then we can list the files in the root directory
	result := &shared.ListToolsResult{}
	options := &shared.RequestOptions{
		// OnProgress: func(progress shared.Progress) {,
	}
	err = client.ListTools(shared.ListToolsRequestParams{}, result, options)
	require.NoError(t, err)
	require.NotEmpty(t, result.Tools)
}
