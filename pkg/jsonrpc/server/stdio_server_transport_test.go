package server

import (
	"bytes"
	"context"
	"encoding/json"
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

	t.Run("should not read until started", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		transport.SetOnError(func(err error) {
			require.NoError(t, err)
		})

		didRead := false
		readMessage := make(chan jsonrpc.JSONRPCMessage, 1)

		transport.SetOnMessage(func(message jsonrpc.JSONRPCMessage) {
			didRead = true
			readMessage <- message
		})

		message := jsonrpc.JSONRPCRequest{
			Jsonrpc: "2.0",
			Id:      jsonrpc.RequestId(1),
			Method:  "ping",
		}
		serialized, err := json.Marshal(message)
		require.NoError(t, err)

		// Push message before the server started
		input.Write(serialized)
		input.Write([]byte("\n"))

		assert.False(t, didRead, "Should not have read message before start")

		err = transport.Start()
		require.NoError(t, err)

		received := <-readMessage
		request, ok := received.(*jsonrpc.JSONRPCRequest)
		require.True(t, ok)
		assert.Equal(t, message, *request)
	})

	t.Run("should read multiple messages", func(t *testing.T) {
		transport := NewStdioServerTransport(ctx, input, output)
		transport.SetOnError(func(err error) {
			require.NoError(t, err)
		})

		messages := []jsonrpc.JSONRPCMessage{
			&jsonrpc.JSONRPCRequest{
				Jsonrpc: "2.0",
				Id:      jsonrpc.RequestId(1),
				Method:  "ping",
			},
			&jsonrpc.JSONRPCNotification{
				Jsonrpc: "2.0",
				Method:  "initialized",
			},
		}

		readMessages := make([]jsonrpc.JSONRPCMessage, 0)
		finished := make(chan struct{})

		transport.SetOnMessage(func(message jsonrpc.JSONRPCMessage) {
			readMessages = append(readMessages, message)
			if len(readMessages) == len(messages) {
				close(finished)
			}
		})

		// Push both messages before starting the server
		for _, m := range messages {
			serialized, err := json.Marshal(m)
			require.NoError(t, err)
			input.Write(serialized)
			input.Write([]byte("\n"))
		}

		// when we start the server transport
		err := transport.Start()

		// then we should have read both messages
		require.NoError(t, err)
		<-finished
		assert.Equal(t, messages, readMessages)
	})
}
