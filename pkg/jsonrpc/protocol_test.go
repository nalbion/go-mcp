package jsonrpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRemoveResponseHandler(t *testing.T) {
	// given
	ctx := context.Background()
	p := NewProtocol(ctx)
	p.requestHandlers = map[Method]RequestHandler{
		"foo": func(ctx context.Context, request *JSONRPCRequest, extra RequestHandlerExtra) (Result, error) {
			return Result{}, nil
		},
	}
	require.NotEmpty(t, p.requestHandlers)

	transport := &TransportBase{}
	p.Connect(ctx, transport)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// a response will be received
	go func() {
		defer wg.Done()

		time.Sleep(1 * time.Millisecond)
		transport.OnMessage(&JSONRPCResponse{
			Id: 1,
			Result: Result{
				AdditionalProperties: map[string]interface{}{
					"foo": "bar",
				},
			},
		})
	}()

	// placeholder result to be unmarshalled into
	result := Result{
		AdditionalProperties: map[string]interface{}{},
	}

	// when we call SendRequest `progressHandlers`
	err := p.SendRequest(ctx, "test", nil, &result)
	require.NoError(t, err)

	wg.Wait()

	// then the result is populated with the Result
	require.Equal(t, map[string]interface{}{"foo": "bar"}, result.AdditionalProperties)
}
