package jsonrpc

// Describes the minimal contract for a MCP transport that a client or server can communicate over.

type Transport interface {
	// called by Protocol.Connect()
	Start() error
	Send(message JSONRPCMessage) error
	Close() error

	SetOnClose(func())
	SetOnError(func(err error))
	SetOnMessage(func(message JSONRPCMessage))
}

type TransportBase struct {
	OnClose   func()
	OnError   func(err error)
	OnMessage func(message JSONRPCMessage)
}

func (t *TransportBase) SetOnClose(f func()) {
	t.OnClose = f
}

func (t *TransportBase) SetOnError(f func(err error)) {
	t.OnError = f
}

func (t *TransportBase) SetOnMessage(f func(message JSONRPCMessage)) {
	t.OnMessage = f
}

func (t *TransportBase) Close() error {
	if t.OnClose != nil {
		t.OnClose()
	}
	return nil
}

// Starts processing messages on the transport, including any connection steps that might need to be taken.
// This method should only be called after callbacks are installed, or else messages may be lost.
//
// NOTE: This method should not be called explicitly when using Client, Server, or Protocol classes,
// as they will implicitly call start().
func (t *TransportBase) Start() error {
	return nil
}

func (t *TransportBase) Send(message JSONRPCMessage) error {
	return nil
}

// type StdioTransport struct{}

// type HTTPTransport struct{}
