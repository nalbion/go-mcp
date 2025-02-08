package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadBuffer(t *testing.T) {
	ctx := context.Background()
	testMessage := JSONRPCRequest{Method: "foobar"}

	t.Run("should have no messages after initialization", func(t *testing.T) {
		readBuffer := NewReadBuffer(ctx)
		message, err := readBuffer.ReadMessage()

		require.ErrorIs(t, err, io.EOF)
		require.Nil(t, message)
	})

	t.Run("should only yield a message after a newline", func(t *testing.T) {
		// given ReadBuffer that we can write to
		// reader, writer := io.Pipe()
		readBuffer := NewReadBuffer(ctx)

		// we write a message without \n
		messageBytes, err := json.Marshal(testMessage)
		require.NoError(t, err)
		readBuffer.Append(messageBytes)
		require.NoError(t, err)

		// when we send the \n
		readBuffer.Append([]byte("\n"))

		// then the message is available
		message, err := readBuffer.ReadMessage()
		require.NoError(t, err)
		require.NotNil(t, message)
		require.Equal(t, "foobar", message.(*JSONRPCRequest).Method)
	})

	t.Run("should be reusable after clearing", func(t *testing.T) {
		// given ReadBuffer that we can write to
		readBuffer := NewReadBuffer(ctx)

		// we write a message without \n
		readBuffer.Append([]byte("foobar"))

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
		readBuffer.Append(messageBytes)

		// then the message is available
		message, err = readBuffer.ReadMessage()
		require.NoError(t, err)
		require.NotNil(t, message)
		require.Equal(t, "foobar", message.(*JSONRPCRequest).Method)
	})
}
