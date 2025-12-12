package infrastructure

import (
	"log"
	"os"
)

type Logger interface {
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
}

type SimpleLogger struct {
	logger *log.Logger
}

func NewLogger() Logger {
	return &SimpleLogger{
		logger: log.New(os.Stderr, "[mcp-k8s-server] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *SimpleLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Printf("[INFO] %s %v", msg, keyvals)
}

func (l *SimpleLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Printf("[ERROR] %s %v", msg, keyvals)
}

func (l *SimpleLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Printf("[DEBUG] %s %v", msg, keyvals)
}

func (l *SimpleLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Printf("[WARN] %s %v", msg, keyvals)
}
