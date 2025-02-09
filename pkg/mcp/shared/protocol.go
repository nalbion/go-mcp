package shared

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp"
)

const DEFAULT_REQUEST_TIMEOUT = 1 * time.Minute

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
	progressHandlers        map[int]mcp.ProgressHandler
	requestAbortControllers sync.Map
}

func NewProtocol(ctx context.Context, options *ProtocolOptions) *Protocol {
	p := &Protocol{
		Protocol: *jsonrpc.NewProtocol(ctx),
	}

	p.Protocol.OnRequest = p.onRequest
	p.Protocol.RemoveResponseHandler = p.removeResponseHandler

	p.Protocol.SetNotificationHandler(NotificationsCancelledMethod, func(notification *jsonrpc.JSONRPCNotification) error {
		if cancelled, ok := notification.Params.AdditionalProperties.(mcp.CancelledNotificationParams); ok {
			p.cancelRequest(cancelled.RequestId)
			return nil
		}

		return fmt.Errorf("failed to assert notification as CancelledNotification")
	})

	return p
}

func (p *Protocol) Connect(ctx context.Context, transport jsonrpc.Transport) error {
	p.Protocol.SetNotificationHandler(NotificationsProgressMethod, func(notification *jsonrpc.JSONRPCNotification) error {
		if progress, ok := notification.Params.AdditionalProperties.(mcp.ProgressNotificationParams); ok {
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

func (p *Protocol) SendRequest(
	ctx context.Context,
	method jsonrpc.Method,
	params *jsonrpc.JSONRPCRequestParams,
	result *jsonrpc.Result,
	options *mcp.RequestOptions,
) error {
	if !p.Protocol.IsConnected() {
		return errors.New("not connected")
	}

	if p.options != nil && p.options.EnforceStrictCapabilities {
		p.Protocol.AssertCapabilityForMethod(method)
	}

	jsonrpcRequest, messageID := p.Protocol.NewRequest(method, params)

	if options != nil && options.OnProgress != nil {
		p.progressHandlers[messageID] = options.OnProgress
		if jsonrpcRequest.Params == nil {
			jsonrpcRequest.Params = &jsonrpc.JSONRPCRequestParams{}
		}
		if jsonrpcRequest.Params.Meta == nil {
			jsonrpcRequest.Params.Meta = &jsonrpc.JSONRPCRequestParamsMeta{}
		}
		progressToken := mcp.ProgressToken(messageID)
		// If specified, the caller is requesting out-of-band progress notifications for
		// this request (as represented by notifications/progress). The value of this
		// parameter is an opaque token that will be attached to any subsequent
		// notifications. The receiver is not obligated to provide these notifications.
		(*jsonrpcRequest.Params.Meta)["progressToken"] = progressToken
	}

	timeout := DEFAULT_REQUEST_TIMEOUT
	if options != nil && options.Timeout > 0 {
		timeout = options.Timeout
	}
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)

	return p.Protocol.SendRequestInternal(ctx, jsonrpcRequest, messageID, result, cancelTimeout, func(reason string) {
		delete(p.progressHandlers, messageID)

		if p.Protocol.IsConnected() {
			err := p.Protocol.SendNotification(
				NotificationsCancelledMethod,
				&jsonrpc.JSONRPCNotificationParams{
					AdditionalProperties: map[string]any{
						"requestId": messageID,
						"reason":    reason,
					},
				},
			)
			if err != nil {
				p.Protocol.OnError(fmt.Errorf("failed to send cancel notification: %v", err))
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

func (p *Protocol) onProgress(notification mcp.ProgressNotificationParams) error {
	progressToken := notification.ProgressToken

	handler := p.progressHandlers[int(progressToken)]
	if handler == nil {
		p.Protocol.OnError(fmt.Errorf("received a progress notification for an unknown token: %v", progressToken))
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
