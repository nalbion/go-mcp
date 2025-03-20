package shared

import (
	"log"
	"os"
)

// MCPLogger defines the interface for logging in the MCP implementation
type MCPLogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// StdLogger is a simple implementation of the MCPLogger interface using the standard log package
type StdLogger struct {
	logger *log.Logger
}

func (l *StdLogger) Debug(format string, args ...interface{}) {
	l.logger.Printf("DEBUG: "+format, args...)
}

func (l *StdLogger) Info(format string, args ...interface{}) {
	l.logger.Printf("INFO: "+format, args...)
}

func (l *StdLogger) Warn(format string, args ...interface{}) {
	l.logger.Printf("WARN: "+format, args...)
}

func (l *StdLogger) Error(format string, args ...interface{}) {
	l.logger.Printf("ERROR: "+format, args...)
}

// DefaultLogger is the default implementation of the MCPLogger interface
var DefaultLogger MCPLogger = &StdLogger{log.New(os.Stdout, "mcp: ", log.LstdFlags)}

// Logger is the legacy logger for backward compatibility
var Logger = log.New(os.Stdout, "mcp: ", log.LstdFlags)
