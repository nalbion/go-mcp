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
	jsonrpc.TransportBase

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

func NewSSEClientTransport(client *http.Client, urlString string, reconnectionTime time.Duration, requestBuilder func(req *http.Request)) (*SSEClientTransport, error) {
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	return &SSEClientTransport{
		client:           client,
		url:              parsedUrl,
		reconnectionTime: reconnectionTime,
		requestBuilder:   requestBuilder,
		endpoint:         make(chan string, 1),
		job:              &sync.WaitGroup{},
	}, nil
}

func (s *SSEClientTransport) Start(ctx context.Context) error {
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
			case <-ctx.Done():
				return
			case event := <-s.session.Incoming():
				switch string(event.Event) {
				case "error":
					err := fmt.Errorf("error receiving SSE error: %s", event.Data)
					s.TransportBase.OnError(err)
					return
				case "open":
					// The connection is open, but we need to wait for the endpoint to be received.
				case "endpoint":
					endpointUrl, err := url.Parse(s.baseUrl + "/" + string(event.Data))
					if err != nil {
						s.TransportBase.OnError(err)
						s.Close()
						return
					}
					s.endpoint <- endpointUrl.String()
				default:
					var message jsonrpc.JSONRPCMessage
					if err := json.Unmarshal([]byte(event.Data), &message); err != nil {
						s.TransportBase.OnError(err)
					} else {
						s.TransportBase.OnMessage(message)
					}
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.endpoint:
		return nil
	}
}

func (s *SSEClientTransport) Send(ctx context.Context, message jsonrpc.JSONRPCMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case endpoint := <-s.endpoint:
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		s.requestBuilder(req)

		resp, err := s.client.Do(req)
		if err != nil {
			s.TransportBase.OnError(err)
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
	s.TransportBase.OnClose()
	s.job.Wait()
	return nil
}

func (s *SSEClientTransport) OnClose(block func()) {
	old := s.TransportBase.OnClose
	s.TransportBase.OnClose = func() {
		old()
		block()
	}
}

func (s *SSEClientTransport) OnError(block func(err error)) {
	old := s.TransportBase.OnError
	s.TransportBase.OnError = func(err error) {
		old(err)
		block(err)
	}
}

func (s *SSEClientTransport) OnMessage(block func(message jsonrpc.JSONRPCMessage)) {
	old := s.TransportBase.OnMessage
	s.TransportBase.OnMessage = func(message jsonrpc.JSONRPCMessage) {
		old(message)
		block(message)
	}
}
