package shared

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/stretchr/testify/require"
)

func TestRemoveResponseHandler(t *testing.T) {
	// given
	ctx := context.Background()
	p := NewProtocol(ctx, &ProtocolOptions{})
	// messageReceived := false
	p.progressHandlers = map[int]ProgressHandler{
		1: func(progress ProgressNotificationParams) {
			// messageReceived = true
		},
	}
	require.NotEmpty(t, p.progressHandlers)

	transport := &jsonrpc.BaseTransport{}
	p.Connect(ctx, transport)
	// client, server := jsonrpc.NewClientServerInMemoryTransports()
	// transort := &jsonrpc.TransportBase{}
	// p.Connect(ctx, client)
	// serverP := NewProtocol(ctx, &ProtocolOptions{})
	// serverP.Connect(ctx, server)
	// server.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)

	// when we receive a response
	go func() {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)
		transport.OnMessage(&jsonrpc.JSONRPCResponse{
			Id: 1,
			Result: jsonrpc.Result{
				AdditionalProperties: map[string]interface{}{
					"foo": "bar",
				},
			},
		})
	}()

	// when we call SendRequest `progressHandlers`
	result := jsonrpc.Result{
		AdditionalProperties: map[string]interface{}{},
	}
	err := p.SendRequest(ctx, "test", nil, &result, nil)
	require.NoError(t, err)

	// p.RemoveResponseHandler(1)

	wg.Wait()

	// then the response and progress handler should be removed
	require.Empty(t, p.progressHandlers)
	// because we removed them when the message was received
	// require.True(t, messageReceived)
}
