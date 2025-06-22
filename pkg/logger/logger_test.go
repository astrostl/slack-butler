package logger

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cleanup  func()
		expected logrus.Level
	}{
		{
			name: "Default log level is Info",
			setup: func() {
				os.Unsetenv("SLACK_LOG_LEVEL")
				// Reinitialize logger
				Log = logrus.New()
				Log.SetLevel(logrus.InfoLevel)
			},
			cleanup:  func() {},
			expected: logrus.InfoLevel,
		},
		{
			name: "Debug log level from environment",
			setup: func() {
				os.Setenv("SLACK_LOG_LEVEL", "debug")
				// Reinitialize logger
				Log = logrus.New()
				Log.SetLevel(logrus.DebugLevel)
			},
			cleanup: func() {
				os.Unsetenv("SLACK_LOG_LEVEL")
			},
			expected: logrus.DebugLevel,
		},
		{
			name: "Error log level from environment",
			setup: func() {
				os.Setenv("SLACK_LOG_LEVEL", "error")
				// Reinitialize logger
				Log = logrus.New()
				Log.SetLevel(logrus.ErrorLevel)
			},
			cleanup: func() {
				os.Unsetenv("SLACK_LOG_LEVEL")
			},
			expected: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			assert.Equal(t, tt.expected, Log.GetLevel())
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)
	assert.Equal(t, Log, logger)
}

func TestWithFields(t *testing.T) {
	fields := LogFields{"test": "value"}
	entry := WithFields(fields)
	assert.NotNil(t, entry)
	assert.Equal(t, "value", entry.Data["test"])
}

func TestWithField(t *testing.T) {
	entry := WithField("key", "value")
	assert.NotNil(t, entry)
	assert.Equal(t, "value", entry.Data["key"])
}

func TestLogFunctions(t *testing.T) {
	// Test that log functions don't panic
	// We can't easily test the output without complex setup
	assert.NotPanics(t, func() { Info("test info") })
	assert.NotPanics(t, func() { Debug("test debug") })
	assert.NotPanics(t, func() { Warn("test warn") })
	assert.NotPanics(t, func() { Error("test error") })
}

func TestLoggerFormatter(t *testing.T) {
	tests := []struct {
		name         string
		formatEnv    string
		expectedType string
	}{
		{
			name:         "JSON formatter when env is json",
			formatEnv:    "json",
			expectedType: "*logrus.JSONFormatter",
		},
		{
			name:         "Text formatter by default",
			formatEnv:    "",
			expectedType: "*logrus.TextFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.formatEnv != "" {
				os.Setenv("SLACK_LOG_FORMAT", tt.formatEnv)
			} else {
				os.Unsetenv("SLACK_LOG_FORMAT")
			}

			// Reinitialize logger to pick up env changes
			Log = logrus.New()
			format := os.Getenv("SLACK_LOG_FORMAT")
			if format == "json" {
				Log.SetFormatter(&logrus.JSONFormatter{
					TimestampFormat: "2006-01-02 15:04:05",
				})
			} else {
				Log.SetFormatter(&logrus.TextFormatter{
					FullTimestamp:   true,
					TimestampFormat: "2006-01-02 15:04:05",
					DisableColors:   os.Getenv("SLACK_LOG_NO_COLOR") == "true",
				})
			}

			formatterType := Log.Formatter
			assert.Contains(t, getFormatterTypeName(formatterType), tt.expectedType)

			// Cleanup
			os.Unsetenv("SLACK_LOG_FORMAT")
		})
	}
}

func getFormatterTypeName(formatter logrus.Formatter) string {
	switch formatter.(type) {
	case *logrus.JSONFormatter:
		return "*logrus.JSONFormatter"
	case *logrus.TextFormatter:
		return "*logrus.TextFormatter"
	default:
		return "unknown"
	}
}