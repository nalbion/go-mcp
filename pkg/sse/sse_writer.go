package sse

import (
	"encoding/json"
	"io"
)

// SSEWriter wraps an io.Writer to help write sse formated data.
type SSEWriter struct {
	writer io.Writer
}

func NewSSEWriter(w io.Writer) *SSEWriter {
	return &SSEWriter{
		writer: w,
	}
}

// writeSSEDone writes a [DONE] SSE message to the writer.
func (w *SSEWriter) WriteDone() {
	_, _ = w.writer.Write([]byte("data: [DONE]\n\n"))
}

// writeSSEData writes a data SSE message to the writer.
func (w *SSEWriter) WriteJsonData(data any) error {
	_, _ = w.writer.Write([]byte("data: "))
	if err := json.NewEncoder(w.writer).Encode(data); err != nil {
		return err
	}

	_, _ = w.writer.Write([]byte("\n")) // Encode() adds one newline, so add only one more here.
	return nil
}

// writeSSEEvent writes a data SSE message to the writer.
func (w *SSEWriter) WriteEvent(name string) error {
	_, err := w.writer.Write([]byte("event: " + name))
	if err != nil {
		return err
	}
	_, err = w.writer.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}
