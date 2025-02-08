package jsonrpc

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
)

// ReadBuffer buffers a continuous stdio stream into discrete JSON-RPC messages.
// used by StdIOClientTransport and StdIOServerTransport
type ReadBuffer struct {
	ctx       context.Context
	reader    *bufio.Reader
	onMessage func(JSONRPCMessage)
	onError   func(error)
}

func NewReadBuffer(
	ctx context.Context,
	reader io.Reader,
	onMessage func(JSONRPCMessage),
	onError func(error),
) *ReadBuffer {
	return &ReadBuffer{
		ctx:       ctx,
		reader:    bufio.NewReader(reader),
		onMessage: onMessage,
		onError:   onError,
	}
}

func (rb *ReadBuffer) Close() {
	rb.reader.Reset(nil)
	rb.reader = nil
}

func (rb *ReadBuffer) Clear() {
	rb.reader.Reset(rb.reader)
}

func (rb *ReadBuffer) Start() {
	for {
		select {
		case <-rb.ctx.Done():
			rb.Close()
			return
		default:
			message, err := rb.ReadMessage()
			if err != nil {
				if err != io.EOF {
					Logger.Printf("failed to receive message: %s", err)
					if rb.onError != nil {
						rb.onError(err)
					}
				}
				return
			}

			if rb.onMessage != nil {
				rb.onMessage(message)
			}
		}
	}
}

func (rb *ReadBuffer) Append(chunk []byte) {
	// rb.buffer.Write(chunk)
}

func (rb *ReadBuffer) ReadMessage() (JSONRPCMessage, error) {
	var contentLength int64
	var content []byte

	for {
		if rb.reader == nil {
			return nil, errors.New("read buffer has been closed")
		}

		header, err := rb.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				Logger.Println("closed connection")
			} else {
				Logger.Printf("failed to read header: %s\n", err)
			}
			return nil, err
		}

		if len(header) <= 2 {
			// empty line before the JSON response
			if contentLength == 0 {
				continue
			}
			break
		}

		if header[0] == '{' {
			content = []byte(header)
			break
		}

		if !strings.HasPrefix(header, "Content-Length: ") {
			// some servers send multiple headers, Content-Type is officially supported by LSP
			continue
		}

		contentLength, err = strconv.ParseInt(header[16:len(header)-2], 10, 32)
		if err != nil {
			Logger.Printf("failed to parse Content-Length: %s\n", err)
			return nil, err
		}

		if contentLength == 0 {
			Logger.Println("empty message")
			continue
		}
	}

	if content == nil {
		content = make([]byte, contentLength)
		_, err := io.ReadFull(rb.reader, content)
		if err != nil {
			return nil, err
		}
	}

	return ParseJSONRPCMessage(content)
}
