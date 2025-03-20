package jsonrpc

import "sync"

// Mock implementations for testing
type MockTransport struct {
	SentRequests      []*JSONRPCRequest
	SentNotifications []*JSONRPCNotification
	// requestHandler      func(*JSONRPCRequest) (Result, error)
	// notificationHandler func(*JSONRPCNotification) error
	mu sync.Mutex
}

func (m *MockTransport) Start() error {
	return nil
}

func (m *MockTransport) Close() error {
	return nil
}

func (m *MockTransport) Send(message JSONRPCMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if notification, ok := message.(*JSONRPCNotification); ok {
		m.SentNotifications = append(m.SentNotifications, notification)
	}

	// m.sentNotifications = append(m.sentNotifications, message)
	return nil
}

// func (m *MockTransport) SendRequest(request *JSONRPCRequest) (Result, error) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.sentRequests = append(m.sentRequests, request)
// 	if m.requestHandler != nil {
// 		return m.requestHandler(request)
// 	}
// 	return Result{}, nil
// }

// func (m *MockTransport) SendNotification(notification *JSONRPCNotification) error {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.sentNotifications = append(m.sentNotifications, notification)
// 	if m.notificationHandler != nil {
// 		return m.notificationHandler(notification)
// 	}
// 	return nil
// }

// func (m *MockTransport) SetRequestHandler(method Method, handler RequestHandler) {
// 	// Not needed for testing
// }

// func (m *MockTransport) SetNotificationHandler(method Method, handler NotificationHandler) {
// 	// Not needed for testing
// }

func (m *MockTransport) SetOnClose(f func()) {
	// Not needed for testing
}

func (m *MockTransport) SetOnError(f func(err error)) {
	// Not needed for testing
}

func (m *MockTransport) SetOnMessage(f func(message JSONRPCMessage)) {
	// Not needed for testing
}
