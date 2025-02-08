package jsonrpc

import (
	"errors"
	"sync"
)

func NewClientServerInMemoryTransports() (*InMemoryTransport, *InMemoryTransport) {
	clientTransport := newInMemoryTransport()
	serverTransport := newInMemoryTransport()

	clientTransport.otherTransport = serverTransport
	serverTransport.otherTransport = clientTransport

	return clientTransport, serverTransport
}

func newInMemoryTransport() *InMemoryTransport {
	return &InMemoryTransport{
		messageQueue: make([]JSONRPCMessage, 0),
		// messageQueue: make(chan(JSONRPCMessage))
	}
}

type InMemoryTransport struct {
	BaseTransport
	otherTransport *InMemoryTransport
	// messageQueue  chan (JSONRPCMessage)
	messageQueue []JSONRPCMessage
	mu           sync.Mutex
}

func (t *InMemoryTransport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Process any messages that were queued before start was called
	for len(t.messageQueue) > 0 {
		message := t.messageQueue[0]
		t.messageQueue = t.messageQueue[1:]
		if t.OnMessage != nil {
			t.OnMessage(message)
		}
	}

	// go func() {
	// 	for message := range reqChan {
	// 		resChan <- message
	// 		if t.OnMessage != nil {
	// 			t.OnMessage(message)
	// 		}
	// 	}
	// }()

	return nil
}

func (t *InMemoryTransport) Send(message JSONRPCMessage) error {
	if t.otherTransport == nil {
		return errors.New("not connected")
	}

	if t.otherTransport.OnMessage != nil {
		t.otherTransport.OnMessage(message)
	} else {
		t.otherTransport.messageQueue = append(t.otherTransport.messageQueue, message)
	}

	return nil
}

func (t *InMemoryTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	other := t.otherTransport
	t.otherTransport = nil
	if other != nil {
		other.Close()
	}
	if t.OnClose != nil {
		t.OnClose()
	}

	return nil
}
