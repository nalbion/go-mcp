package shared

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
)

const DEFAULT_REQUEST_TIMEOUT = 1 * time.Minute

type (
	ProgressHandler func(progress ProgressNotificationParams)
)

type ProtocolOptions struct {
	// Whether to restrict emitted requests to only those that the remote side has indicated that they can handle, through their advertised capabilities.
	// Note that this DOES NOT affect checking of _local_ side capabilities, as it is considered a logic error to mis-specify those.
	// Currently this defaults to false, for backwards compatibility with SDK versions that did not advertise capabilities correctly. In future, this will default to true.
	EnforceStrictCapabilities bool
	Timeout                   time.Duration
}

type Protocol struct {
	jsonrpc.Protocol
	options                 *ProtocolOptions
	progressHandlers        map[int]ProgressHandler
	requestAbortControllers sync.Map
}

func NewProtocol(ctx context.Context, options *ProtocolOptions) *Protocol {
	p := &Protocol{
		Protocol: *jsonrpc.NewProtocol(ctx),
	}

	p.OnRequest = p.onRequest
	p.Protocol.RemoveResponseHandler = p.removeResponseHandler

	p.SetNotificationHandler(NotificationsCancelledMethod, func(notification *jsonrpc.JSONRPCNotification) error {
		if cancelled, ok := notification.Params.AdditionalProperties.(CancelledNotificationParams); ok {
			p.cancelRequest(cancelled.RequestId)
			return nil
		}

		return fmt.Errorf("failed to assert notification as CancelledNotification")
	})

	return p
}

func (p *Protocol) Connect(ctx context.Context, transport jsonrpc.Transport) error {
	p.SetNotificationHandler(NotificationsProgressMethod, func(notification *jsonrpc.JSONRPCNotification) error {
		if progress, ok := notification.Params.AdditionalProperties.(ProgressNotificationParams); ok {
			return p.onProgress(progress)
		}

		return fmt.Errorf("failed to assert notification as ProgressNotification")
	})

	err := p.Protocol.Connect(ctx, transport)
	if err != nil {
		return err
	}

	transport.SetOnClose(p.onClose)

	return nil
}

type RequestOptions struct {
	// If set, requests progress notifications from the remote end (if supported).
	// When progress notifications are received, this callback will be invoked.
	onProgress ProgressHandler
	// Can be used to cancel an in-flight request. This will cause an context.Canceled to be returned from SendRequest().
	cancel context.CancelFunc
	// A timeout for this request. If exceeded, an McpError with code `RequestTimeout` will be returned from SendRequest().
	// If not specified, `DEFAULT_REQUEST_TIMEOUT` will be used as the timeout.
	timeout time.Duration
}

func (p *Protocol) SendRequest(
	ctx context.Context,
	method jsonrpc.Method,
	params *jsonrpc.JSONRPCRequestParams,
	result *jsonrpc.Result,
	options *RequestOptions,
) error {
	if !p.IsConnected() {
		return errors.New("not connected")
	}

	if p.options != nil && p.options.EnforceStrictCapabilities {
		p.AssertCapabilityForMethod(method)
	}

	jsonrpcRequest, messageID := p.NewRequest(method, params)

	if options != nil && options.onProgress != nil {
		p.progressHandlers[messageID] = options.onProgress
		if jsonrpcRequest.Params == nil {
			jsonrpcRequest.Params = &jsonrpc.JSONRPCRequestParams{}
		}
		if jsonrpcRequest.Params.Meta == nil {
			jsonrpcRequest.Params.Meta = &jsonrpc.JSONRPCRequestParamsMeta{}
		}
		progressToken := ProgressToken(messageID)
		// If specified, the caller is requesting out-of-band progress notifications for
		// this request (as represented by notifications/progress). The value of this
		// parameter is an opaque token that will be attached to any subsequent
		// notifications. The receiver is not obligated to provide these notifications.
		(*jsonrpcRequest.Params.Meta)["progressToken"] = progressToken
	}

	timeout := DEFAULT_REQUEST_TIMEOUT
	if options != nil && options.timeout > 0 {
		timeout = options.timeout
	}
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)

	return p.SendRequestInternal(ctx, jsonrpcRequest, messageID, result, cancelTimeout, func(reason string) {
		delete(p.progressHandlers, messageID)

		if p.IsConnected() {
			err := p.SendNotification(
				NotificationsCancelledMethod,
				&jsonrpc.JSONRPCNotificationParams{
					AdditionalProperties: map[string]any{
						"requestId": messageID,
						"reason":    reason,
					},
				},
			)
			if err != nil {
				p.OnError(fmt.Errorf("failed to send cancel notification: %v", err))
			}
		}
	})
}

// ctx is required to maintain consistency with the jsonrpc.Protocol.OnRequest()
// which receives the cancelable context created here.
func (p *Protocol) onRequest(ctx context.Context, request *jsonrpc.JSONRPCRequest, onDone func()) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p.requestAbortControllers.Store(request.Id, cancel)

	p.Protocol.OnRequest(ctx, request, func() {
		p.requestAbortControllers.Delete(request.Id)
	})
}

func (p *Protocol) cancelRequest(id jsonrpc.RequestId) {
	if cancel, ok := p.requestAbortControllers.Load(id); ok {
		cancel.(context.CancelFunc)()
	}
}

func (p *Protocol) onProgress(notification ProgressNotificationParams) error {
	progressToken := notification.ProgressToken

	handler := p.progressHandlers[int(progressToken)]
	if handler == nil {
		p.OnError(fmt.Errorf("received a progress notification for an unknown token: %v", progressToken))
		return nil
	}

	handler(notification)
	return nil
}

func (p *Protocol) removeResponseHandler(id int) {
	delete(p.progressHandlers, id)
	// p.Protocol.RemoveResponseHandler(id)
}

func (p *Protocol) onClose() {
	for k := range p.progressHandlers {
		delete(p.progressHandlers, k)
	}

	p.Protocol.Close()
}
