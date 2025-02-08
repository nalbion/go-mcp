package sse

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/r3labs/sse"
)

type ClientSSESession struct {
	client  *sse.Client
	events  chan *sse.Event
	context context.Context
	cancel  context.CancelFunc
}

func NewClientSSESession(client *http.Client, url *url.URL, reconnectionTime time.Duration, requestBuilder func(req *http.Request)) *ClientSSESession {
	ctx, cancel := context.WithCancel(context.Background())
	sseClient := sse.NewClient(url.String())
	sseClient.Connection = client
	// sseClient.ReconnectStrategy = reconnectionTime

	session := &ClientSSESession{
		client:  sseClient,
		events:  make(chan *sse.Event),
		context: ctx,
		cancel:  cancel,
	}

	go session.listen()
	return session
}

func (s *ClientSSESession) listen() {
	s.client.SubscribeWithContext(s.context, "", func(event *sse.Event) {
		s.events <- event
	})
}

func (s *ClientSSESession) Incoming() <-chan *sse.Event {
	return s.events
}

func (s *ClientSSESession) Close() {
	s.cancel()
	close(s.events)
}
