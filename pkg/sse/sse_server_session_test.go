package sse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerSSESession_Send(t *testing.T) {
	t.Run("should return error if writer is not initialized", func(t *testing.T) {
		session := &ServerSSESession{}
		err := session.Send(NewServerSentEvent().WithData("test"))
		assert.EqualError(t, err, "writer not initialized")
	})

	t.Run("should send single-line data", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		// when
		err := session.Send(NewServerSentEvent().WithData("Hello, world!"))

		// then
		require.NoError(t, err)
		assert.Equal(t, "data: Hello, world!\n\n", session.String())
	})

	t.Run("should send multi-line data", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		err := session.Send(NewServerSentEvent().WithData("Hello,\nworld!"))
		require.NoError(t, err)

		assert.Equal(t, "data: Hello,\ndata: world!\n\n", session.String())
	})

	t.Run("should send JSON data spread across multiple lines", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		jsonData := `{"key1": "value1", "key2": "value2"}`

		// when
		err := session.Send(NewServerSentEvent().WithData(jsonData))

		// then
		require.NoError(t, err)
		expected := "data: {\"key1\": \"value1\", \"key2\": \"value2\"}\n\n"
		assert.Equal(t, expected, session.String())
	})

	t.Run("should send event with ID", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		eventID := "12345"

		// when
		err := session.Send(NewServerSentEvent().WithID(eventID).WithData("Hello, world!"))

		// then
		require.NoError(t, err)
		expected := "id: 12345\ndata: Hello, world!\n\n"
		assert.Equal(t, expected, session.String())
	})

	t.Run("should send event with retry/reconnection-timeout", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		retry := 3000

		// when
		err := session.Send(NewServerSentEvent().WithRetry(retry).WithData("Hello, world!"))

		// then
		require.NoError(t, err)
		expected := "retry: 3000\ndata: Hello, world!\n\n"
		assert.Equal(t, expected, session.String())
	})

	t.Run("should send event with event name", func(t *testing.T) {
		session, _ := NewServerSSESession(&SSESessionOptions{Buffered: true})
		defer session.Close()

		eventName := "greeting"

		// when
		err := session.Send(NewServerSentEvent().WithEvent(eventName).WithData("Hello, world!"))

		// then
		require.NoError(t, err)
		expected := "event: greeting\ndata: Hello, world!\n\n"
		assert.Equal(t, expected, session.String())
	})
}
