package mcp

import (
	"context"
	"time"
)

type (
	ProgressHandler func(progress ProgressNotificationParams)
)

type RequestOptions struct {
	// If set, requests progress notifications from the remote end (if supported).
	// When progress notifications are received, this callback will be invoked.
	OnProgress ProgressHandler
	// Can be used to Cancel an in-flight request. This will cause an context.Canceled to be returned from SendRequest().
	Cancel context.CancelFunc
	// A Timeout for this request. If exceeded, an McpError with code `RequestTimeout` will be returned from SendRequest().
	// If not specified, `DEFAULT_REQUEST_TIMEOUT` will be used as the Timeout.
	Timeout time.Duration
}
