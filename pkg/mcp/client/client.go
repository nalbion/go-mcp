package client

import (
	"context"
	"fmt"
	"slices"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp/shared"
)

type ClientOptions struct {
	// The Capabilities this client supports
	Capabilities shared.ClientCapabilities
	// Whether to strictly enforce capabilities when interacting with the server
	// defaults to true
	EnforceStrictCapabilities *bool
}

// An MCP client on top of a pluggable transport.
// The client automatically performs the initialization handshake with the server when Connect() is called.
// After initialization, [severCapabilities] and [serverVersion] provide details about the connected server.
//
// You can extend this class with custom request/notification/result types if needed.
//
// @param clientInfo Information about the client implementation (name, version).
// @param options Configuration options for this client.
type Client struct {
	*shared.Protocol
	ctx          context.Context
	clientInfo   shared.Implementation
	capabilities shared.ClientCapabilities
	// after the initialization process completes, this will contain the server's capabilities
	ServerCapabilities *shared.ServerCapabilities
	ServerVersion      string
}

func NewClient(
	ctx context.Context,
	clientInfo shared.Implementation,
	options ClientOptions,
) *Client {
	enforceStrictCapabilities := true
	if options.EnforceStrictCapabilities != nil {
		enforceStrictCapabilities = *options.EnforceStrictCapabilities
	}

	c := &Client{
		Protocol: shared.NewProtocol(
			ctx,
			&shared.ProtocolOptions{
				EnforceStrictCapabilities: enforceStrictCapabilities,
			},
		),
		ctx:          ctx,
		clientInfo:   clientInfo,
		capabilities: options.Capabilities,
	}

	// c.Protocol.SetContext(ctx)
	// c.Protocol.EnforceStrictCapabilities = enforceStrictCapabilities

	return c
}

func (c *Client) Connect(transport jsonrpc.Transport) error {
	if err := c.Protocol.Connect(c.ctx, transport); err != nil {
		return err
	}

	connected := false
	defer func() {
		if !connected {
			c.Close()
		}
	}()

	result := &shared.InitializeResult{}
	err := c.SendRequest(
		shared.InitializeMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: shared.InitializeRequestParams{
				ProtocolVersion: shared.LatestProtocolVersion,
				Capabilities:    c.capabilities,
				ClientInfo:      c.clientInfo,
			},
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		nil)
	if err != nil {
		return err
	}

	if !slices.Contains(shared.SupportedProtocolVersions, result.ProtocolVersion) {
		return fmt.Errorf("server's protocol version is not supported: %s", result.ProtocolVersion)
	}

	shared.Logger.Printf("Connected to server: %v\n", result)

	c.ServerCapabilities = &result.Capabilities
	c.ServerVersion = result.ServerInfo.Version

	err = c.SendNotification(shared.NotificationsInitializedMethod, nil)
	if err != nil {
		return err
	}

	return nil
}

// Ping() sends a ping request to the server to check connectivity.
func (c *Client) Ping(options *shared.RequestOptions) error {
	return c.SendRequest(shared.PingMethod, nil, nil, options)
}

// Complate() sends a completion request to the server, typically to generate or complete some content
// returns the completion result returned by the server, or `null` if none.
func (c *Client) Complete(params shared.CompleteRequestParams, result *shared.CompleteResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.CompletionCompleteMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

// SetLogggingLevel() sets the logging level on the server.
func (c *Client) SetLogggingLevel(level shared.LoggingLevel, options *shared.RequestOptions) error {
	return c.SendRequest(shared.LoggingSetLevelMethod, &jsonrpc.JSONRPCRequestParams{
		AdditionalProperties: shared.SetLevelRequestParams{
			Level: level,
		},
	}, nil, options)
}

// Lists all available prompts from the server.
func (c *Client) ListPrompts(params shared.ListPromptsRequestParams, result *shared.ListPromptsResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ListPromptsMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

// GetPrompt() retrieves a prompt by name from the server.
// returns the requested prompt details, or `null` if not found
func (c *Client) GetPrompt(params shared.GetPromptRequestParams, result *shared.GetPromptResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.GetPromptsMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) ListResources(params shared.ListResourcesRequestParams, result *shared.ListResourcesResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ListResourcesMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) ListResourceTemplates(params shared.ListResourceTemplatesRequestParams, result *shared.ListResourceTemplatesResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ListResourcesTemplatesMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) ReadResource(params shared.ReadResourceRequestParams, result *shared.ReadResourceResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ReadResourcesMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) SubscribeResources(params shared.SubscribeRequestParams, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ResourcesSubscribeMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		nil,
		options)
}

func (c *Client) UnsubscribeResources(params shared.UnsubscribeRequestParams, options *shared.RequestOptions) error {
	return c.SendRequest(shared.ResourcesUnsubscribeMethod, &jsonrpc.JSONRPCRequestParams{
		AdditionalProperties: params,
	}, nil, options)
}

func (c *Client) ListTools(params shared.ListToolsRequestParams, result *shared.ListToolsResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ToolsListMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) CallTool(params shared.CallToolRequestParams, result *shared.CallToolResult, options *shared.RequestOptions) error {
	return c.SendRequest(
		shared.ToolsCallMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)
}

func (c *Client) SendRootsListChangedNotification(params shared.RootsListChangedNotification) error {
	return c.SendNotification(shared.NotificationsRootsListChangedMethod, &jsonrpc.JSONRPCNotificationParams{
		AdditionalProperties: params,
	})
}

func (c *Client) AssertCapability(capability string, method string) error {
	caps := c.ServerCapabilities

	switch capability {
	case "logging":
		if caps.Logging != nil {
			return nil
		}
	case "prompts":
		if caps.Prompts != nil {
			return nil
		}
	case "resources":
		if caps.Resources != nil {
			return nil
		}
	case "tools":
		if caps.Tools != nil {
			return nil
		}
	}

	return fmt.Errorf("server does not support %s (required for %s)", capability, method)
}

func (c *Client) AssertCapabilityForMethod(method jsonrpc.Method) error {
	switch method {
	case shared.LoggingSetLevelMethod:
		if c.ServerCapabilities.Logging == nil {
			return fmt.Errorf("server does not support logging (required for %s)", method)
		}
	case shared.GetPromptsMethod,
		shared.ListPromptsMethod,
		shared.CompletionCompleteMethod:
		if c.ServerCapabilities.Prompts == nil {
			return fmt.Errorf("server does not support prompts (required for %s)", method)
		}
	case shared.ListResourcesMethod,
		shared.ListResourcesTemplatesMethod,
		shared.ReadResourcesMethod,
		shared.ResourcesSubscribeMethod,
		shared.ResourcesUnsubscribeMethod:
		if c.ServerCapabilities.Resources == nil {
			return fmt.Errorf("server does not support resources (required for %s)", method)
		}

		if method == shared.ResourcesSubscribeMethod && (c.ServerCapabilities.Resources.Subscribe == nil || !*c.ServerCapabilities.Resources.Subscribe) {
			return fmt.Errorf("server does not support resource subscriptions (required for %s)", method)
		}
	case shared.ToolsListMethod,
		shared.ToolsCallMethod:
		if c.ServerCapabilities.Tools == nil {
			return fmt.Errorf("server does not support tools (required for %s)", method)
		}
	}

	return nil
}

func (c *Client) AssertNotificationCapability(method jsonrpc.Method) error {
	switch method {
	case shared.NotificationsRootsListChangedMethod:
		if c.capabilities.Roots == nil || c.capabilities.Roots.ListChanged == nil || !*c.capabilities.Roots.ListChanged {
			return fmt.Errorf("client does not support roots list changed notifications (required for %s)", method)
		}
	}

	return nil
}

func (c *Client) AssertRequestHandlerCapability(method jsonrpc.Method) error {
	switch method {
	case shared.SamplingCreateMessageMethod:
		if c.capabilities.Sampling == nil {
			return fmt.Errorf("client does not support sampling capability (required for %s)", method)
		}
	case shared.RootsListMethod:
		if c.capabilities.Roots == nil {
			return fmt.Errorf("client does not support roots capability (required for %s)", method)
		}
	}

	return nil
}

func (c *Client) SendRequest(
	method jsonrpc.Method,
	params *jsonrpc.JSONRPCRequestParams,
	result *jsonrpc.Result,
	options *shared.RequestOptions,
) error {
	return c.Protocol.SendRequest(
		c.ctx,
		method,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		//
		result,
		options)
}

// func RunClient(ctx context.Context, urlOrCommand string, args []string) {
// 	client := NewClient(ctx,
// 		shared.Implementation{
// 			Name:    "mcp test client",
// 			Version: "0.1.0",
// 		},
// 		ClientOptions{
// 			Capabilities: shared.ClientCapabilities{
// 				Sampling: shared.ClientCapabilitiesSampling{},
// 			},
// 		})

// 	var clientTransport shared.Transport

// 	// if urlOrCommand == "" {
// 	// 	var serverTransport shared.Transport
// 	// 	clientTransport, serverTransport = shared.NewClientServerInMemoryTransports()
// 	// 	serverTransport.Start()
// 	// }

// 	parsedURL, err := url.Parse(urlOrCommand)
// 	if err == nil {
// 		switch parsedURL.Scheme {
// 		// case "http", "https":
// 		// 	clientTransport = NewSSEClientTransport(parsedURL)
// 		// case "ws", "wss":
// 		// 	clientTransport = NewWebSocketClientTransport(parsedURL)
// 		default:
// 			clientTransport = NewStdioClientTransport(ctx, StdioServerParameters{
// 				Command: urlOrCommand,
// 				Args:    args,
// 			})
// 		}
// 	} else {
// 		clientTransport = NewStdioClientTransport(ctx, StdioServerParameters{
// 			Command: urlOrCommand,
// 			Args:    args,
// 		})
// 	}

// 	err = client.Connect(clientTransport)
// 	if err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}

// 	fmt.Println("Connected to server.")

// 	// Implement request and close logic here

// 	fmt.Println("Closed.")
// }
