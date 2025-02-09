package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	RequestHandlerExtra any
	ResponseOrError     any
	RequestHandler      func(ctx context.Context, request *JSONRPCRequest, extra RequestHandlerExtra) (Result, error)
	ResponseHandler     func(response *JSONRPCResponse, err error)
	NotificationHandler func(notification *JSONRPCNotification) error
)

type Protocol struct {
	ctx context.Context

	transport            Transport
	requestMessageID     int
	requestHandlers      map[Method]RequestHandler
	notificationHandlers map[Method]NotificationHandler
	responseHandlers     map[int]ResponseHandler

	OnRequest             func(ctx context.Context, request *JSONRPCRequest, onDone func())
	RemoveResponseHandler func(id int)
	// onClose is a callback for when the connection is closed for any reason.
	// This is invoked when close() is called as well.
	onClose func()
	// Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
	onError                     func(err error)
	fallbackRequestHandler      RequestHandler
	fallbackNotificationHandler NotificationHandler
}

func (p *Protocol) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func NewProtocol(ctx context.Context) *Protocol {
	p := &Protocol{
		ctx:                  ctx,
		requestHandlers:      make(map[Method]RequestHandler),
		notificationHandlers: make(map[Method]NotificationHandler),
		responseHandlers:     make(map[int]ResponseHandler),
	}

	p.OnRequest = p.onRequest
	p.RemoveResponseHandler = p.removeResponseHandler

	// P.SetNotificationHandler(ProgressNotificationSchema, (notification) => {
	// 	P.onprogress(notification as unknown as ProgressNotification);
	//   });

	// P.setRequestHandler(PingRequestSchema,
	// 	// Automatic pong by default.
	// 	(_request) => ({}) as SendResultT,
	//   );

	return p
}

// Connect attaches to the given transport, starts it, and starts listening for messages.
// The Protocol object assumes ownership of the Transport, replacing any callbacks that have already been set,
// and expects that it is the only user of the Transport instance going forward.
func (p *Protocol) Connect(ctx context.Context, transport Transport) error {
	p.transport = transport
	p.transport.SetOnClose(p.onCloseImpl)
	p.transport.SetOnError(func(err error) {
		if p.onError != nil {
			p.onError(err)
		}
	})

	p.transport.SetOnMessage(func(message JSONRPCMessage) {
		switch message := message.(type) {
		case JSONRPCRequest:
			p.OnRequest(ctx, &message, nil)
		case JSONRPCResponse:
			p.onResponse(&message, nil)
		case *JSONRPCError:
			p.onResponse(nil, message)
		case JSONRPCNotification:
			p.onNotification(&message)
		default:
			p.OnError(fmt.Errorf("unknown message type: %T", message))
		}
	})

	return p.transport.Start()
}

func (p *Protocol) IsConnected() bool {
	return p.transport != nil
}

func (p *Protocol) Close() error {
	// avoid infinite loop. we tell the transport to call onCloseImpl() from t.Close().
	if p.transport == nil {
		p.transport.Close()
	} else {
		p.onCloseImpl()
	}
	return nil
}

func (p *Protocol) onCloseImpl() {
	responseHandlers := p.responseHandlers
	p.responseHandlers = make(map[int]ResponseHandler)

	p.transport = nil
	if p.onClose != nil {
		p.onClose()
	}

	err := NewJSONRPCErrorError(0, ConnectionClosed, "Connection closed", nil)
	for _, handler := range responseHandlers {
		handler(nil, err)
	}
}

func (p *Protocol) OnError(err error) {
	if p.onError != nil {
		p.onError(err)
	}
}

// Overriden by MCP Client/Server
func (p *Protocol) AssertCapabilityForMethod(method Method) {
}

func (p *Protocol) NewRequest(method Method, params *JSONRPCRequestParams) (*JSONRPCRequest, int) {
	p.requestMessageID++
	messageID := p.requestMessageID

	return &JSONRPCRequest{
		Jsonrpc: "2.0",
		Id:      RequestId(messageID),
		Method:  string(method),
		Params:  params,
	}, messageID
}

// SendRequest sends a request and wait for a response.
// `result` is provided to allow the caller to specify the type of the result in advance.
// Default values can be set in the `result` for any nullable fields.
// Do not use this method to emit notifications! Use SendNotification() instead.
func (p *Protocol) SendRequest(
	ctx context.Context,
	method Method,
	params *JSONRPCRequestParams,
	result *Result,
) error {
	if !p.IsConnected() {
		return errors.New("not connected")
	}

	jsonrpcRequest, messageID := p.NewRequest(method, params)

	return p.SendRequestInternal(ctx, jsonrpcRequest, messageID, result, nil, nil)
}

// this is a "protected" method for use by jsonrpc/mcp.Protocol.SendRequest() only.
func (p *Protocol) SendRequestInternal(
	ctx context.Context,
	jsonrpcRequest *JSONRPCRequest,
	messageID int,
	result *Result,
	cancelTimeout context.CancelFunc,
	onCancel func(reason string),
) error {
	resChan := make(chan *JSONRPCResponse, 1)
	errChan := make(chan error, 1)

	p.responseHandlers[messageID] = func(response *JSONRPCResponse, err error) {
		if err != nil {
			errChan <- err
		} else {
			// if result != nil && result.AdditionalProperties != nil {
			// 	// now we know the result type we want, we can unmarshal the additional properties into it.
			// 	str, err := json.Marshal(response.Result.AdditionalProperties)
			// 	if err != nil {
			// 		errChan <- err
			// 		return
			// 	}
			// 	err = json.Unmarshal(str, result.AdditionalProperties)
			// 	if err != nil {
			// 		errChan <- err
			// 		return
			// 	}
			// 	response.Result.AdditionalProperties = result.AdditionalProperties
			// }
			resChan <- response
		}

		if cancelTimeout != nil {
			Logger.Println("Cancelling timeout")
			cancelTimeout()
		}
	}

	cancel := func(reason string) {
		Logger.Printf("Cancelling request: %s\n", reason)
		delete(p.responseHandlers, messageID)

		if cancelTimeout != nil {
			cancelTimeout()
		}
		if onCancel != nil {
			onCancel(reason)
		}

		// if p.transport != nil {
		// 	errChan <- fmt.Errorf("request cancelled: %s", reason)
		// }
	}

	if err := p.transport.Send(jsonrpcRequest); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			// return nil
		} else {
			cancel("context done")
			return ctx.Err()
		}
	case err := <-errChan:
		if errors.Is(err, context.Canceled) {
			return nil
		} else if errors.Is(err, context.DeadlineExceeded) {
			cancel("request timed out")
			return nil
		}
		return err
	case response := <-resChan:
		return parseResponse(response, result)
	}

	return errors.New("Protocol.SendRequestInternal: unexpected state")
}

// this is a "protected" method for use by jsonrpc/mcp.Protocol.SendRequest() only.
// Use SendRequest() which manages request IDs, response handlers, timeouts etc.
func (p *Protocol) SendInternal(jsonrpcMessage JSONRPCMessage) error {
	if p.transport != nil {
		return p.transport.Send(jsonrpcMessage)
	}
	// maintaining same behavior as the original code and ignoring non-connected state.
	return nil
}

// parseResponse should only be called (and JSONRPCResponse unmarshalled) after verifying that the message is not a JSONRPCError.
// messageResult _may_ be provided if the result type is known in advance.
func parseResponse(response *JSONRPCResponse, messageResult *Result) error {
	content, err := json.Marshal(response.Result.AdditionalProperties)
	if err != nil {
		return err
	}

	if messageResult != nil {
		err = json.Unmarshal(content, messageResult.AdditionalProperties)
		if err != nil {
			return err
		}

		return nil
	}

	err = ParseResult(content, messageResult)
	if err != nil {
		return err
	}

	return nil
}

func (p *Protocol) SendNotification(method Method, params *JSONRPCNotificationParams) error {
	if p.transport == nil {
		return errors.New("not connected")
	}

	// p.assertNotificationCapability(method);

	return p.transport.Send(NewJSONRPCNotification(method, params))
}

func (p *Protocol) onNotification(notification *JSONRPCNotification) error {
	var err error
	if handler, ok := p.notificationHandlers[Method(notification.Method)]; ok {
		err = handler(notification)
	} else if p.fallbackNotificationHandler != nil {
		err = p.fallbackNotificationHandler(notification)
	}

	if err != nil {
		p.OnError(fmt.Errorf("uncaught error in notification handler: %v", err))
	}
	return err
}

func (p *Protocol) SetRequestHandler(method Method, handler RequestHandler) {
	// p.assertRequestHandlerCapability(method);
	p.requestHandlers[method] = handler
}

func (p *Protocol) RemoveRequestHandler(method Method) {
	delete(p.requestHandlers, method)
}

func (p *Protocol) SetNotificationHandler(method Method, handler NotificationHandler) {
	p.notificationHandlers[method] = handler
}

func (p *Protocol) RemoveNotificationHandler(method Method) {
	delete(p.notificationHandlers, method)
}

// onRequest is called by the Transport when a JSONRPCRequest is received.
// mcp.Protocol calls this with a cancelable ctx.
func (p *Protocol) onRequest(ctx context.Context, request *JSONRPCRequest, onDone func()) {
	handler, ok := p.requestHandlers[Method(request.Method)]
	if !ok {
		handler = p.fallbackRequestHandler
	}

	if handler == nil {
		err := p.transport.Send(NewJSONRPCError(
			request.Id,
			JSONRPCErrorError{
				Code:    int(MethodNotFound),
				Message: "Method not found",
			},
		))
		if err != nil {
			p.OnError(fmt.Errorf("failed to send error response: %v", err))
		}
		return
	}

	go func() {
		if onDone != nil {
			defer onDone()
		}

		result, err := handler(ctx, request, nil)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			var jsonrpcErrorError *JSONRPCErrorError
			if !errors.As(err, &jsonrpcErrorError) {
				jsonrpcErrorError = &JSONRPCErrorError{
					Code:    int(InternalError),
					Message: err.Error(),
				}
			}

			if sendErr := p.transport.Send(NewJSONRPCError(
				request.Id,
				*jsonrpcErrorError,
			)); sendErr != nil {
				p.OnError(fmt.Errorf("failed to send error response: %w", sendErr))
			}
			return
		}

		if ctx.Err() == context.Canceled {
			return
		}

		if sendErr := p.transport.Send(newJSONRPCResponse(request.Id, result)); sendErr != nil {
			p.OnError(fmt.Errorf("failed to send response: %w", sendErr))
		}
	}()
}

func (p *Protocol) onResponse(response *JSONRPCResponse, errorResponse *JSONRPCError) {
	var id int
	var result *Result
	var err *JSONRPCErrorError
	if response != nil {
		id = int(response.Id)
		result = &response.Result
	} else if errorResponse != nil {
		id = int(errorResponse.Id)
		err = &errorResponse.Error
	} else {
		p.OnError(fmt.Errorf("invalid response type: %T", response))
		return
	}

	handler, ok := p.responseHandlers[id]
	if !ok {
		p.OnError(fmt.Errorf("received response for unknown request ID: %v", id))
		return
	}

	p.RemoveResponseHandler(id)

	if result != nil {
		handler(response, nil)
	} else {
		handler(nil, err)
	}
}

func (p *Protocol) removeResponseHandler(id int) {
	delete(p.responseHandlers, id)
}

func newJSONRPCResponse(id RequestId, result Result) *JSONRPCResponse {
	return &JSONRPCResponse{
		Id:      id,
		Jsonrpc: "2.0",
		Result:  result,
	}
}

func NewJSONRPCNotification(method Method, params *JSONRPCNotificationParams) *JSONRPCNotification {
	return &JSONRPCNotification{
		Jsonrpc: "2.0",
		Method:  string(method),
		Params:  params,
	}
}

type ErrorCode int

const (
	// SDK error codes
	ConnectionClosed ErrorCode = -32000
	RequestTimeout   ErrorCode = -32001

	// Standard JSON-RPC error codes
	ParseError     ErrorCode = -32700
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
)

func NewJSONRPCErrorError(id RequestId, code ErrorCode, message string, data any) *JSONRPCErrorError {
	return &JSONRPCErrorError{
		Code:    int(code),
		Message: message,
		Data:    data,
	}
}

func (e *JSONRPCErrorError) Error() string {
	return fmt.Sprintf("MCPError %d: %s", e.Code, e.Message)
}

func NewJSONRPCError(id RequestId, err JSONRPCErrorError) *JSONRPCError {
	return &JSONRPCError{
		Id:      id,
		Jsonrpc: "2.0",
		Error:   err,
	}
}

// type AbortController struct {
// 	cancel context.CancelFunc
// }

// func (a *AbortController) Abort() {
// 	a.cancel()
// }

// func newAbortController(ctx context.Context) *AbortController {
// 	ctx, cancel := context.WithCancel(ctx)
// 	return &AbortController{
// 		cancel: cancel,
// 	}
// }
