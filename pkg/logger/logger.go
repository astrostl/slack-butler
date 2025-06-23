package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// LogFields is an alias for logrus.Fields for convenience.
type LogFields = logrus.Fields

var Log *logrus.Logger

func init() {
	Log = logrus.New()

	// Set output to stdout
	Log.SetOutput(os.Stdout)

	// Set log level based on environment (default to Info)
	level := os.Getenv("SLACK_LOG_LEVEL")
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Use JSON formatter for production, text for development
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
}

// GetLogger returns the configured logger instance.
func GetLogger() *logrus.Logger {
	return Log
}

// WithFields creates a new logger entry with structured fields.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// WithField creates a new logger entry with a single structured field.
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// Info logs an info message.
func Info(msg string) {
	Log.Info(msg)
}

// Debug logs a debug message.
func Debug(msg string) {
	Log.Debug(msg)
}

// Warn logs a warning message.
func Warn(msg string) {
	Log.Warn(msg)
}

// Error logs an error message.
func Error(msg string) {
	Log.Error(msg)
}

// Fatal logs a fatal message and exits.
func Fatal(msg string) {
	Log.Fatal(msg)
}
