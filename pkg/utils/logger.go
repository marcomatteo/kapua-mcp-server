package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging capabilities
type Logger struct {
	level  LogLevel
	prefix string
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(prefix string, level LogLevel) *Logger {
	return &Logger{
		level:  level,
		prefix: prefix,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger(prefix string) *Logger {
	level := LogLevelInfo
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		level = parseLogLevel(levelStr)
	}
	return NewLogger(prefix, level)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", format, args...)
	}
}

// log performs the actual logging
func (l *Logger) log(level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.prefix != "" {
		l.logger.Printf("[%s] [%s] %s", level, l.prefix, message)
	} else {
		l.logger.Printf("[%s] %s", level, message)
	}
}

// parseLogLevel converts a string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// FormatError creates a formatted error message
func FormatError(operation string, err error) string {
	return fmt.Sprintf("%s failed: %v", operation, err)
}

// SafeString safely extracts a string from a potentially nil pointer
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
