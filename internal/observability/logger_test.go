package observability

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		output  string
		wantErr bool
	}{
		{
			name:    "default stdout json",
			level:   "info",
			format:  "json",
			output:  "stdout",
			wantErr: false,
		},
		{
			name:    "console format",
			level:   "debug",
			format:  "console",
			output:  "stdout",
			wantErr: false,
		},
		{
			name:    "stderr output",
			level:   "warn",
			format:  "json",
			output:  "stderr",
			wantErr: false,
		},
		{
			name:    "empty output defaults to stdout",
			level:   "info",
			format:  "json",
			output:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level, tt.format, tt.output)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if logger == nil {
				t.Error("expected logger but got nil")
			}
		})
	}
}

func TestNewLoggerFileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger("info", "json", logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger == nil {
		t.Fatal("expected logger but got nil")
	}

	// Log something
	logger.Info("test message")

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"ERROR", zerolog.ErrorLevel},
		{"unknown", zerolog.InfoLevel},
		{"", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if level != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestLoggerWithRequestID(t *testing.T) {
	logger, err := NewLogger("debug", "json", "stdout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loggerWithReqID := logger.WithRequestID("req-123")
	if loggerWithReqID == nil {
		t.Error("expected logger with request ID but got nil")
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger, err := NewLogger("debug", "json", "stdout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := map[string]interface{}{
		"user_id":  "user-123",
		"endpoint": "/api/test",
	}
	loggerWithFields := logger.WithFields(fields)
	if loggerWithFields == nil {
		t.Error("expected logger with fields but got nil")
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	logger := &Logger{
		logger: zerolog.New(&buf).With().Timestamp().Logger(),
	}

	// Test each log method
	logger.Debug("debug message")
	logger.Debugf("debug %s", "formatted")
	logger.Info("info message")
	logger.Infof("info %s", "formatted")
	logger.Warn("warn message")
	logger.Warnf("warn %s", "formatted")
	logger.Error("error message")
	logger.Errorf("error %s", "formatted")

	output := buf.String()

	// Verify messages were logged
	if !strings.Contains(output, "debug message") {
		t.Error("missing debug message")
	}
	if !strings.Contains(output, "info message") {
		t.Error("missing info message")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("missing warn message")
	}
	if !strings.Contains(output, "error message") {
		t.Error("missing error message")
	}
}

func TestLoggerErrorWithErr(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: zerolog.New(&buf).With().Timestamp().Logger(),
	}

	testErr := os.ErrNotExist
	logger.ErrorWithErr(testErr, "file not found")

	output := buf.String()
	if !strings.Contains(output, "file not found") {
		t.Error("missing error message")
	}
}

func TestGetZerolog(t *testing.T) {
	logger, err := NewLogger("info", "json", "stdout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	zl := logger.GetZerolog()
	// Verify it returns a valid zerolog.Logger
	if zl.GetLevel() < zerolog.TraceLevel {
		t.Error("invalid zerolog logger returned")
	}
}
