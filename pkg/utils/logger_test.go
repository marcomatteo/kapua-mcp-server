package utils

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	cases := map[string]LogLevel{
		"DEBUG":   LogLevelDebug,
		"info":    LogLevelInfo,
		"Warn":    LogLevelWarn,
		"error":   LogLevelError,
		"unknown": LogLevelInfo,
	}
	for input, expected := range cases {
		if got := parseLogLevel(input); got != expected {
			t.Errorf("parseLogLevel(%q) = %v, want %v", input, got, expected)
		}
	}
}

func TestFormatError(t *testing.T) {
	err := errors.New("boom")
	if got := FormatError("op", err); got != "op failed: boom" {
		t.Errorf("unexpected FormatError output: %q", got)
	}
}

func TestSafeString(t *testing.T) {
	if SafeString(nil) != "" {
		t.Errorf("SafeString(nil) expected empty string")
	}
	s := "hello"
	if SafeString(&s) != "hello" {
		t.Errorf("SafeString(&s) expected 'hello'")
	}
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	orig := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = orig
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestLoggerLevels(t *testing.T) {
	output := captureOutput(func() {
		logger := NewLogger("test", LogLevelInfo)
		logger.Debug("debug message")
		logger.Info("info message")
	})
	if strings.Contains(output, "debug message") {
		t.Errorf("debug message should not be logged at info level")
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("info message not logged: %s", output)
	}
}

func TestNewDefaultLoggerEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "error")
	logger := NewDefaultLogger("pref")
	if logger.level != LogLevelError {
		t.Errorf("expected level %v, got %v", LogLevelError, logger.level)
	}
}
