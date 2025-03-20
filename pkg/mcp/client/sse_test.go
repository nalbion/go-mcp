package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	jsonrpcclient "github.com/nalbion/go-mcp/pkg/jsonrpc/client"

	"github.com/nalbion/go-mcp/pkg/mcp"
	"github.com/nalbion/go-mcp/pkg/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEClientTransport(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create an SSE server session
		session, _ := sse.NewServerSSESession(&sse.SSESessionOptions{Buffered: true})

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		// Send a test event
		err := session.Send(sse.NewServerSentEvent().
			WithEvent("message").
			WithData(`{"jsonrpc":"2.0","method":"test","params":{}}`))
		if err != nil {
			t.Logf("Error sending SSE event: %v", err)
		}

		// Keep the connection open
		<-ctx.Done()
	}))
	defer server.Close()

	// Create an SSE client
	client := NewClient(ctx, mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, ClientOptions{})

	// Create an SSE client transport
	transport, err := jsonrpcclient.NewDefaultSSEClientTransport(ctx, server.URL, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a notification handler to verify we receive the test event
	receivedNotification := make(chan bool, 1)
	transport.SetOnMessage(func(message jsonrpc.JSONRPCMessage) {
		if notification, ok := message.(jsonrpc.JSONRPCNotification); ok {
			if notification.Method == "test" {
				receivedNotification <- true
			}
		}
	})

	// when we connect to the SSE server
	err = client.Connect(transport)

	// then the connection is successful
	require.NoError(t, err)

	// Wait for the notification to be received
	select {
	case <-receivedNotification:
		// Success - notification was received
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for SSE notification")
	}
}

func TestSSEClientReconnection(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a counter for connection attempts
	connectionAttempts := 0

	// Create a channel to signal when the server should close the connection
	closeConnection := make(chan struct{})

	// Create a test server that will close the connection after receiving a request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionAttempts++

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		// Flush the headers
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Close the connection after a short delay if requested
		select {
		case <-closeConnection:
			// Connection will be closed when the handler returns
			return
		case <-ctx.Done():
			return
		}
	}))
	defer server.Close()

	// Create an SSE client
	client := NewClient(ctx, mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, ClientOptions{})

	// Create an SSE client transport with a short reconnect delay
	transport, err := jsonrpcclient.NewDefaultSSEClientTransport(ctx, server.URL, 500*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// when we connect to the SSE server
	err = client.Connect(transport)

	// then the connection is successful
	require.NoError(t, err)

	// Initial connection
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 1, connectionAttempts)

	// when we close the connection
	closeConnection <- struct{}{}

	// then the client should reconnect
	time.Sleep(1 * time.Second)
	assert.GreaterOrEqual(t, connectionAttempts, 2, "Expected at least one reconnection attempt")
}
