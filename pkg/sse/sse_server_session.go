package sse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type SSESessionOptions struct {
	// if not buffered, writer and reader will be connected via a pipe
	// and calls to Send() will block until the reader reads the data
	Buffered bool
}

// SSE can be implemented in AWS Lambda:
//
//		func sseHandler(ctx context.Context, evt events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
//		  session, reader := sse.NewServerSSESession(&SSESessionOptions{Buffered: true})
//	   // do something with evt.Body etc
//
//		  session.Send(sse.ServerSentEvent{Data: "Hello, world!"})
//		  return session.CreateLambdaStreamingResponse(), nil
//		}
func NewServerSSESession(options *SSESessionOptions) (*ServerSSESession, io.Reader) {
	var reader io.Reader
	var writer io.Writer
	buffered := options != nil && options.Buffered

	if buffered {
		buf := &bytes.Buffer{}
		reader = buf
		writer = buf
	} else {
		reader, writer = io.Pipe()
	}

	session := &ServerSSESession{
		reader:   reader,
		writer:   writer,
		buffered: buffered,
	}

	return session, reader
}

type ServerSSESession struct {
	reader   io.Reader
	writer   io.Writer
	buffered bool
}

// see https://web.dev/articles/eventsource-basics#event_stream_format
func (s *ServerSSESession) Send(event *ServerSentEvent) error {
	if s.writer == nil {
		return errors.New("writer not initialized")
	}

	if event.Comments != nil {
		if _, err := s.writer.Write([]byte(fmt.Sprintf(":%s\n", *event.Comments))); err != nil {
			return err
		}
	}

	if event.ID != nil {
		if _, err := s.writer.Write([]byte(fmt.Sprintf("id: %s\n", *event.ID))); err != nil {
			return err
		}
	}

	if event.Retry != nil {
		if _, err := s.writer.Write([]byte(fmt.Sprintf("retry: %d\n", *event.Retry))); err != nil {
			return err
		}
	}

	if event.Event != nil {
		if _, err := s.writer.Write([]byte(fmt.Sprintf("event: %s\n", *event.Event))); err != nil {
			return err
		}
	}

	if event.Data != nil {
		for _, line := range bytes.Split([]byte(*event.Data), []byte("\n")) {
			if _, err := s.writer.Write([]byte(fmt.Sprintf("data: %s\n", line))); err != nil {
				return err
			}
		}
	}

	if _, err := s.writer.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

func (s *ServerSSESession) CreateLambdaStreamingResponse() *events.LambdaFunctionURLStreamingResponse {
	return &events.LambdaFunctionURLStreamingResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":  "text/event-stream",
			"Connection":    "keep-alive",
			"Cache-Control": "no-cache",
		},
		Body: s.reader,
	}
}

func (s *ServerSSESession) String() string {
	if s.buffered {
		return s.reader.(*bytes.Buffer).String()
	}

	return "[not buffered]"
}

func (s *ServerSSESession) Close() error {
	s.writer = nil
	return nil
}
