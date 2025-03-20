package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/sse"
)

// Client transport for SSE: this will connect to a server using Server-Sent Events for receiving
// messages and make separate POST requests for sending messages.
type SSEClientTransport struct {
	jsonrpc.BaseTransport

	ctx              context.Context
	client           *http.Client
	url              *url.URL
	reconnectionTime time.Duration
	requestBuilder   func(req *http.Request)

	initiated bool
	session   sse.ClientSSESession
	job       *sync.WaitGroup

	endpoint chan string
	baseUrl  string
}

func NewDefaultSSEClientTransport(ctx context.Context, url string, reconnectionTime time.Duration) (*SSEClientTransport, error) {
	client := http.DefaultClient
	requestBuilder := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
	}

	return NewSSEClientTransport(ctx, client, url, reconnectionTime, requestBuilder)
}

func NewSSEClientTransport(ctx context.Context, client *http.Client, urlString string, reconnectionTime time.Duration, requestBuilder func(req *http.Request)) (*SSEClientTransport, error) {
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	return &SSEClientTransport{
		ctx:              ctx,
		client:           client,
		url:              parsedUrl,
		reconnectionTime: reconnectionTime,
		requestBuilder:   requestBuilder,
		endpoint:         make(chan string, 1),
		job:              &sync.WaitGroup{},
	}, nil
}

func (s *SSEClientTransport) Start() error {
	if s.initiated {
		return errors.New("SSEClientTransport already started")
	}
	s.initiated = true

	// s.session = NewClientSSESession(s.client, s.url, s.reconnectionTime, s.requestBuilder)
	s.baseUrl = s.url.String()

	s.job.Add(1)
	go func() {
		defer s.job.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case event := <-s.session.Incoming():
				switch string(event.Event) {
				case "error":
					err := fmt.Errorf("error receiving SSE error: %s", event.Data)
					s.BaseTransport.OnError(err)
					return
				case "open":
					// The connection is open, but we need to wait for the endpoint to be received.
				case "endpoint":
					endpointUrl, err := url.Parse(s.baseUrl + "/" + string(event.Data))
					if err != nil {
						s.BaseTransport.OnError(err)
						s.Close()
						return
					}
					s.endpoint <- endpointUrl.String()
				default:
					var message jsonrpc.JSONRPCMessage
					if err := json.Unmarshal([]byte(event.Data), &message); err != nil {
						s.BaseTransport.OnError(err)
					} else {
						s.BaseTransport.OnMessage(message)
					}
				}
			}
		}
	}()

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case <-s.endpoint:
		return nil
	}
}

func (s *SSEClientTransport) Send(message jsonrpc.JSONRPCMessage) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case endpoint := <-s.endpoint:
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		s.requestBuilder(req)

		resp, err := s.client.Do(req)
		if err != nil {
			s.BaseTransport.OnError(err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.New("Error POSTing to endpoint: " + resp.Status)
		}
		return nil
	}
}

func (s *SSEClientTransport) Close() error {
	if !s.initiated {
		return errors.New("SSEClientTransport is not initialized")
	}

	s.session.Close()
	s.BaseTransport.OnClose()
	s.job.Wait()
	return nil
}

func (s *SSEClientTransport) OnClose(block func()) {
	old := s.BaseTransport.OnClose
	s.BaseTransport.OnClose = func() {
		old()
		block()
	}
}

func (s *SSEClientTransport) OnError(block func(err error)) {
	old := s.BaseTransport.OnError
	s.BaseTransport.OnError = func(err error) {
		old(err)
		block(err)
	}
}

func (s *SSEClientTransport) OnMessage(block func(message jsonrpc.JSONRPCMessage)) {
	old := s.BaseTransport.OnMessage
	s.BaseTransport.OnMessage = func(message jsonrpc.JSONRPCMessage) {
		old(message)
		block(message)
	}
}
