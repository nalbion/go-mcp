package server

import (
	"bytes"
	"context"
	"testing"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioServerTransport(t *testing.T) {
	ctx := context.Background()
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	t.Run("Start", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		err := transport.Start()

		require.NoError(t, err)
		assert.True(t, transport.initialized)
	})

	t.Run("Start_AlreadyStarted", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		err := transport.Start()
		require.NoError(t, err)

		err = transport.Start()
		assert.EqualError(t, err, "StdioServerTransport already started")
	})

	t.Run("Close", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		err := transport.Start()
		require.NoError(t, err)

		err = transport.Close()
		require.NoError(t, err)
		assert.False(t, transport.initialized)
	})

	t.Run("Send", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		err := transport.Start()
		require.NoError(t, err)

		message := &jsonrpc.JSONRPCRequest{
			Jsonrpc: "2.0",
			Id:      jsonrpc.RequestId(1),
			Method:  "testMethod",
			Params: &jsonrpc.JSONRPCRequestParams{
				AdditionalProperties: map[string]interface{}{"param1": "value1"},
			},
		}

		err = transport.Send(message)
		require.NoError(t, err)

		expectedOutput := `{"id":1,"jsonrpc":"2.0","method":"testMethod","params":{"param1":"value1"}}`
		assert.Equal(t, expectedOutput, output.String())
	})
}
