package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp/shared"
)

type StdioServerTransport struct {
	jsonrpc.TransportBase
	ctx          context.Context
	inputStream  io.Reader
	outputStream io.Writer
	readBuffer   shared.ReadBuffer
	initialized  bool
	readingJob   chan struct{}
	readChannel  chan []byte
	outputWriter *bufio.Writer
	lock         sync.Mutex
}

func NewStdioServerTransport(ctx context.Context, inputStream io.Reader, outputStream io.Writer) *StdioServerTransport {
	return &StdioServerTransport{
		ctx:          ctx,
		inputStream:  inputStream,
		outputStream: outputStream,
		readChannel:  make(chan []byte, 100),
		outputWriter: bufio.NewWriter(outputStream),
	}
}

func (s *StdioServerTransport) Start() error {
	if s.initialized {
		return errors.New("StdioServerTransport already started")
	}
	s.initialized = true
	s.readingJob = make(chan struct{})

	go s.readFromStdin()
	go s.processMessages()

	return nil
}

func (s *StdioServerTransport) readFromStdin() {
	buf := make([]byte, 8192)
	for {
		n, err := s.inputStream.Read(buf)
		if err != nil {
			if err != io.EOF {
				if s.OnError != nil {
					s.OnError(err)
				}
			}
			close(s.readingJob)
			return
		}
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			s.readChannel <- chunk
		}
	}
}

func (s *StdioServerTransport) processMessages() {
	for chunk := range s.readChannel {
		s.readBuffer.Append(chunk)
		s.processReadBuffer()
	}
}

func (s *StdioServerTransport) processReadBuffer() {
	for {
		message, err := s.readBuffer.ReadMessage()
		if err != nil {
			if s.OnError != nil {
				s.OnError(err)
			}
			return
		}
		if message == nil {
			break
		}
		if s.OnMessage != nil {
			s.OnMessage(message)
		}
	}
}

func (s *StdioServerTransport) Close() error {
	if !s.initialized {
		return nil
	}
	s.initialized = false
	close(s.readingJob)
	close(s.readChannel)
	s.readBuffer.Clear()
	if s.OnClose != nil {
		s.OnClose()
	}
	return nil
}

func (s *StdioServerTransport) Send(message *jsonrpc.JSONRPCMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err = s.outputWriter.Write(data)
	if err != nil {
		return err
	}
	return s.outputWriter.Flush()
}
