package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp/shared"
)

type StdioServerParameters struct {
	Command string
	Args    []string
	Env     map[string]string
	StdErr  io.Writer
}

type StdioClientTransport struct {
	jsonrpc.BaseTransport
	ctx          context.Context
	cancel       context.CancelFunc
	serverParams StdioServerParameters
	process      *os.Process
	sendChannel  chan jsonrpc.JSONRPCMessage
	readBuffer   *jsonrpc.ReadBuffer
}

func NewStdioClientTransport(ctx context.Context, server StdioServerParameters) *StdioClientTransport {
	return &StdioClientTransport{
		ctx:          ctx,
		serverParams: server,
	}
}

func (t *StdioClientTransport) Start() error {
	if t.process != nil {
		return errors.New("already started! If using Client class, note that Connect() calls Start() automatically")
	}

	ctx := t.ctx
	cmd := exec.Command(t.serverParams.Command, t.serverParams.Args...)
	// if t.serverParams.Env != nil {
	//   cmd.Env = t.serverParams.Env
	// }
	// TODO: filter by DEFAULT_INHERITED_ENV_VARS
	cmd.Env = os.Environ()
	if t.serverParams.StdErr != nil {
		cmd.Stdout = t.serverParams.StdErr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	t.sendChannel = make(chan jsonrpc.JSONRPCMessage)
	go func() {
		for {
			select {
			case <-ctx.Done():
				stdin.Close()
				return
			case message := <-t.sendChannel:
				jsonMessage, err := json.Marshal(message)
				if err != nil {
					if t.OnError != nil {
						t.OnError(err)
					}
					continue
				}

				// jsonrpc.Logger.Printf("sending message: %s\n", string(jsonMessage))
				_, err = io.WriteString(stdin, string(jsonMessage)+"\n")
				if err != nil {
					if t.OnError != nil {
						t.OnError(err)
					}
				}
			}
		}
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		if t.OnError != nil {
			t.OnError(err)
		}
		return fmt.Errorf("failet to start command '%s' with args %v: %w", t.serverParams.Command, t.serverParams.Args, err)
	}

	t.process = cmd.Process
	t.readBuffer = jsonrpc.NewReadBuffer(ctx)

	t.ctx, t.cancel = context.WithCancel(ctx)
	go t.readServerOutput(stdout)
	go t.readServerErr(stderr)

	return nil
}

func (t *StdioClientTransport) Send(message jsonrpc.JSONRPCMessage) error {
	if t.process == nil {
		return errors.New("transport not started")
	}

	t.sendChannel <- message
	return nil
}

func (t *StdioClientTransport) Close() error {
	t.cancel()
	if t.process != nil {
		t.process.Kill()
		t.process.Wait()
		t.process = nil
	}
	if t.readBuffer != nil {
		t.readBuffer.Close()
	}
	return nil
}

func (t *StdioClientTransport) readServerOutput(reader io.Reader) {
	buf := make([]byte, 8192)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				if t.OnError != nil {
					t.OnError(err)
				}
			}
			break
		}
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			t.readBuffer.Append(chunk)
			t.processReadBuffer()
		}
	}
}

func (t *StdioClientTransport) readServerErr(reader io.Reader) {
	buf := make([]byte, 8192)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				if t.OnError != nil {
					t.OnError(err)
				}
			}
			break
		}
		if n > 0 {
			// docker logs progress (and also "daemon not started") to stderr
			shared.Logger.Printf("stderr: %s\n", string(buf[:n]))
		}
	}
}

func (t *StdioClientTransport) processReadBuffer() {
	for {
		message, err := t.readBuffer.ReadMessage()
		if err != nil {
			if t.OnError != nil {
				t.OnError(err)
			}
			return
		}
		if message == nil {
			break
		}
		if t.OnMessage != nil {
			t.OnMessage(message)
		}
	}
}
