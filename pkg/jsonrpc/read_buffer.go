package jsonrpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
)

// ReadBuffer buffers a continuous stdio stream into discrete JSON-RPC messages.
// used by StdIOClientTransport and StdIOServerTransport
type ReadBuffer struct {
	ctx    context.Context
	buffer *bytes.Buffer
}

func NewReadBuffer(
	ctx context.Context,
) *ReadBuffer {
	return &ReadBuffer{
		ctx:    ctx,
		buffer: &bytes.Buffer{},
	}
}

func (rb *ReadBuffer) Close() {
	rb.buffer.Reset()
	rb.buffer = nil
}

func (rb *ReadBuffer) Clear() {
	rb.buffer.Reset()
}

func (rb *ReadBuffer) Append(chunk []byte) {
	rb.buffer.Write(chunk)
}

func (rb *ReadBuffer) ReadMessage() (JSONRPCMessage, error) {
	var contentLength int64
	var content []byte

	for {
		if rb.buffer == nil {
			return nil, errors.New("read buffer has been closed")
		}

		header, err := rb.buffer.ReadString('\n')
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
		_, err := io.ReadFull(rb.buffer, content)
		if err != nil {
			return nil, err
		}
	}

	return ParseJSONRPCMessage(content)
}
