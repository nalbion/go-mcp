package client

import (
	"context"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	ybjsonrpc "github.com/ybbus/jsonrpc"
)

type HttpClientTransport struct {
	jsonrpc.TransportBase
	// endpoint string // eg: "http://my-rpc-service:8080/rpc"
	ctx    context.Context
	client ybjsonrpc.RPCClient
}

func NewHttpClientTransport(ctx context.Context, endpoint string) *HttpClientTransport {
	return &HttpClientTransport{
		ctx:    ctx,
		client: ybjsonrpc.NewClient(endpoint),
	}
}

func (c *HttpClientTransport) Start() error {
	return nil
}

// the response is parsed by Protocol.SendRequest() calling parseResponse()
func (c *HttpClientTransport) Send(message ybjsonrpc.RPCRequest) error {
	response, err := c.client.Call(message.Method, message.Params)
	if err != nil {
		return err
	}

	if response.Error != nil {
		c.OnError(jsonrpc.NewJSONRPCErrorError(
			jsonrpc.RequestId(response.ID),
			jsonrpc.ErrorCode(response.Error.Code),
			response.Error.Message,
			response.Error.Data,
		))
	} else {
		responseMessage := jsonrpc.JSONRPCResponse{
			Id: jsonrpc.RequestId(response.ID),
			Result: jsonrpc.Result{
				AdditionalProperties: response.Result,
			},
		}

		c.OnMessage(responseMessage)
	}

	return nil
}

func (c *HttpClientTransport) Close() error {
	return nil
}
