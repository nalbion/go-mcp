package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadBuffer(t *testing.T) {
	ctx := context.Background()
	testMessage := jsonrpc.JSONRPCRequest{Method: "foobar"}

	t.Run("should have no messages after initialization", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte{})
		readBuffer := NewReadBuffer(ctx, reader, nil, nil)
		message, err := readBuffer.ReadMessage()

		require.ErrorIs(t, err, io.EOF)
		require.Nil(t, message)
	})

	t.Run("should only yield a message after a newline", func(t *testing.T) {
		// given ReadBuffer that we can write to
		// reader, writer := io.Pipe()
		reader := bytes.NewBuffer([]byte{})
		writer := reader
		readBuffer := NewReadBuffer(ctx, reader, nil, nil)

		// we write a message without \n
		messageBytes, err := json.Marshal(testMessage)
		require.NoError(t, err)
		_, err = writer.Write(messageBytes)
		require.NoError(t, err)

		// This test was adapted from https://github.com/modelcontextprotocol/kotlin-sdk/blob/8426228bf8bc205b73f703fd1dbc109830be214b/src/jvmTest/kotlin/shared/ReadBufferTest.kt
		// In go, reader.ReadString() moves the cursor. If this functionality is required, readMessage would have to Peek(lots) and check for \n
		// // when we read
		// message, err := readBuffer.readMessage()

		// // then there is no message available until \n
		// require.ErrorIs(t, err, io.EOF)
		// require.Nil(t, message)

		// when we send the \n
		writer.Write([]byte("\n"))

		// then the message is available
		message, err := readBuffer.ReadMessage()
		require.NoError(t, err)
		require.NotNil(t, message)
		require.Equal(t, "foobar", message.(*jsonrpc.JSONRPCRequest).Method)
	})

	t.Run("should be reusable after clearing", func(t *testing.T) {
		// given ReadBuffer that we can write to
		// reader, writer := io.Pipe()
		reader := bytes.NewBuffer([]byte{})
		writer := reader
		readBuffer := NewReadBuffer(ctx, reader, nil, nil)

		// we write a message without \n
		writer.Write([]byte("foobar"))

		// when we clear the buffer
		readBuffer.Clear()
		message, err := readBuffer.ReadMessage()

		// then there is no message or error
		assert.Nil(t, message)
		assert.ErrorIs(t, err, io.EOF)

		// now, when we write a message with \n
		messageBytes, err := json.Marshal(testMessage)
		require.NoError(t, err)
		messageBytes = append(messageBytes, '\n')
		_, err = writer.Write(messageBytes)

		// then the message is available
		require.NoError(t, err)
		message, err = readBuffer.ReadMessage()
		require.NoError(t, err)
		require.NotNil(t, message)
		require.Equal(t, "foobar", message.(*jsonrpc.JSONRPCRequest).Method)
	})
}
