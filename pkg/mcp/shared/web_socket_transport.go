package shared

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// const MCP_SUBPROTOCOL = "mcp"

// type JSONRPCMessage struct {
// 	// Define the structure of your JSONRPCMessage here
// }

// type WebSocketMcpTransport struct {
// 	client         *http.Client
// 	url            string
// 	requestBuilder func(req *http.Request)
// 	conn           *websocket.Conn
// 	mu             sync.Mutex
// 	initialized    bool

// 	onClose   func()
// 	onError   func(err error)
// 	onMessage func(message JSONRPCMessage)
// }

// func NewWebSocketMcpTransport(client *http.Client, url string, requestBuilder func(req *http.Request)) *WebSocketMcpTransport {
// 	return &WebSocketMcpTransport{
// 		client:         client,
// 		url:            url,
// 		requestBuilder: requestBuilder,
// 	}
// }

// func (w *WebSocketMcpTransport) InitializeSession() error {
// 	headers := http.Header{}
// 	headers.Set("Sec-WebSocket-Protocol", MCP_SUBPROTOCOL)

// 	dialer := websocket.Dialer{
// 		Proxy:            http.ProxyFromEnvironment,
// 		HandshakeTimeout: 45 * time.Second,
// 	}

// 	conn, _, err := dialer.Dial(w.url, headers)
// 	if err != nil {
// 		return err
// 	}

// 	w.conn = conn
// 	return nil
// }

// func (w *WebSocketMcpTransport) Start(ctx context.Context) error {
// 	w.mu.Lock()
// 	if w.initialized {
// 		w.mu.Unlock()
// 		return errors.New("WebSocketClientTransport already started")
// 	}
// 	w.initialized = true
// 	w.mu.Unlock()

// 	if err := w.InitializeSession(); err != nil {
// 		return err
// 	}

// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 				_, message, err := w.conn.ReadMessage()
// 				if err != nil {
// 					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
// 						w.onClose()
// 					} else {
// 						w.onError(err)
// 					}
// 					return
// 				}

// 				var jsonMessage JSONRPCMessage
// 				if err := json.Unmarshal(message, &jsonMessage); err != nil {
// 					w.onError(err)
// 					return
// 				}

// 				w.onMessage(jsonMessage)
// 			}
// 		}
// 	}()

// 	return nil
// }

// func (w *WebSocketMcpTransport) Send(message JSONRPCMessage) error {
// 	w.mu.Lock()
// 	defer w.mu.Unlock()

// 	if !w.initialized {
// 		return errors.New("Not connected")
// 	}

// 	data, err := json.Marshal(message)
// 	if err != nil {
// 		return err
// 	}

// 	return w.conn.WriteMessage(websocket.TextMessage, data)
// }

// func (w *WebSocketMcpTransport) Close() error {
// 	w.mu.Lock()
// 	defer w.mu.Unlock()

// 	if !w.initialized {
// 		return errors.New("Not connected")
// 	}

// 	err := w.conn.Close()
// 	if err != nil {
// 		return err
// 	}

// 	w.onClose()
// 	return nil
// }

// func (w *WebSocketMcpTransport) OnClose(block func()) {
// 	w.onClose = block
// }

// func (w *WebSocketMcpTransport) OnError(block func(err error)) {
// 	w.onError = block
// }

// func (w *WebSocketMcpTransport) OnMessage(block func(message JSONRPCMessage)) {
// 	w.onMessage = block
// }
