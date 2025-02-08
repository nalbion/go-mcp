package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/sse"
)

const SESSION_ID_PARAM = "sessionId"

type SSEServerTransport struct {
	jsonrpc.TransportBase
	ctx         context.Context
	cancel      context.CancelFunc
	endpoint    string
	session     *sse.ServerSSESession
	initialized bool
	sessionId   string
	mu          sync.Mutex
}

func NewSSEServerTransport(ctx context.Context, endpoint string, session *sse.ServerSSESession) *SSEServerTransport {
	ctx, cancel := context.WithCancel(ctx)

	return &SSEServerTransport{
		ctx:         ctx,
		cancel:      cancel,
		endpoint:    endpoint,
		session:     session,
		sessionId:   uuid.New().String(),
		initialized: false,
	}
}

// Handles the initial SSE connection request
// This should be called when a GET request is made to establish the SSE stream
func (s *SSEServerTransport) Start() error {
	s.mu.Lock()
	if s.initialized {
		s.mu.Unlock()
		return errors.New("SSEServerTransport already started")
	}
	s.initialized = true
	s.mu.Unlock()

	data := fmt.Sprintf("%s?%s=%s", s.endpoint, SESSION_ID_PARAM, s.sessionId)
	if err := s.session.Send(
		sse.NewServerSentEvent().
			WithEvent("endpoint").
			WithData(data),
	); err != nil {
		return err
	}

	<-s.ctx.Done()
	if s.OnClose != nil {
		s.OnClose()
	}
	return nil
}

// HandlePostMessage handles POST requests to the SSE endpoint.
// This assumes that the body of the POST request is a JSONRPC message.
func (s *SSEServerTransport) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if !s.initialized {
		s.mu.Unlock()
		http.Error(w, "SSE connection not established", http.StatusInternalServerError)
		if s.OnError != nil {
			s.OnError(errors.New("SSE connection not established"))
		}
		return
	}
	s.mu.Unlock()

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported content-type", http.StatusBadRequest)
		if s.OnError != nil {
			s.OnError(errors.New("unsupported content-type"))
		}
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid message: %v", err), http.StatusBadRequest)
		if s.OnError != nil {
			s.OnError(err)
		}
		return
	}

	if err := s.handleMessage(body); err != nil {
		http.Error(w, fmt.Sprintf("Error handling message %s: %v", body, err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted"))
}

func (s *SSEServerTransport) HandleLambdaRequest(event events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
	if event.Headers["Content-Type"] != "application/json" {
		return &events.LambdaFunctionURLStreamingResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}

	if err := s.handleMessage([]byte(event.Body)); err != nil {
		return &events.LambdaFunctionURLStreamingResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	return s.session.CreateLambdaStreamingResponse(), nil
}

// Handle a client message, regardless of how it arrived.
// This can be used to inform the server of messages that arrive via a means different from HTTP POST.
func (s *SSEServerTransport) handleMessage(message []byte) error {
	parsedMessage, err := jsonrpc.ParseJSONRPCMessage(message)
	if err != nil {
		if s.OnError != nil {
			s.OnError(err)
		}
		return err
	}

	if s.OnMessage != nil {
		s.OnMessage(parsedMessage)
	}
	return nil
}

func (s *SSEServerTransport) Close() error {
	s.cancel()
	if err := s.session.Close(); err != nil {
		return err
	}
	if s.OnClose != nil {
		s.OnClose()
	}
	return nil
}

func (s *SSEServerTransport) Send(message jsonrpc.JSONRPCMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.initialized {
		return errors.New("not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return s.session.Send(sse.NewServerSentEvent().
		WithEvent("message").
		WithData(string(data)))
}
