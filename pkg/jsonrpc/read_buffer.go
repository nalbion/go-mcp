package jsonrpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

		line, err := rb.buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) != 0 {
					// there is more to come
					rb.buffer.WriteString(line)
					// } else {
					// 	Logger.Println("ReadMessage() reached EOF")
				}

				return nil, nil
			}
			return nil, fmt.Errorf("failed to read header: %s", err)
		}

		if len(line) <= 2 {
			// empty line before the JSON response
			if contentLength == 0 {
				continue
			}
			break
		}

		if line[0] == '{' {
			content = []byte(line)
			break
		}

		if !strings.HasPrefix(line, "Content-Length: ") {
			// some servers send multiple headers, Content-Type is officially supported by LSP
			continue
		}

		contentLength, err = strconv.ParseInt(line[16:len(line)-2], 10, 32)
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
