package logrus

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestHook struct {
	Entries []*logrus.Entry
}

func (hook *TestHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *TestHook) Fire(entry *logrus.Entry) error {
	hook.Entries = append(hook.Entries, entry)
	return nil
}

func TestNew(t *testing.T) {
	logger, err := New("production")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.IsType(t, &LogrusProvider{}, logger)
	assert.IsType(t, &logrus.JSONFormatter{}, logger.Formatter)
	assert.Equal(t, logrus.InfoLevel, logger.Level)

	logger, err = New("development")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.TextFormatter{}, logger.Formatter)
	assert.Equal(t, logrus.DebugLevel, logger.Level)
}

func TestLogrusProviderMethods(t *testing.T) {
	logger, _ := New("development")

	// Add a test hook to capture log entries
	hook := &TestHook{}
	logger.AddHook(hook)

	// Test Debug
	logger.Debug("debug message", map[string]any{"key": "value"})
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.DebugLevel, hook.Entries[0].Level)
	assert.Equal(t, "debug message", hook.Entries[0].Message)
	assert.Equal(t, "value", hook.Entries[0].Data["key"])

	// Clear entries
	hook.Entries = nil

	// Test Info
	logger.Info("info message", map[string]any{"key": "value"})
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.Entries[0].Level)
	assert.Equal(t, "info message", hook.Entries[0].Message)
	assert.Equal(t, "value", hook.Entries[0].Data["key"])

	// Clear entries
	hook.Entries = nil

	// Test Warn
	logger.Warn("warn message", map[string]any{"key": "value"})
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.WarnLevel, hook.Entries[0].Level)
	assert.Equal(t, "warn message", hook.Entries[0].Message)

	// Clear entries
	hook.Entries = nil

	// Test Error
	logger.Error("error message", map[string]any{"key": "value"})
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.Entries[0].Level)
	assert.Equal(t, "error message", hook.Entries[0].Message)
	assert.Equal(t, "value", hook.Entries[0].Data["key"])
}

func TestLogrusProviderPanic(t *testing.T) {
	logger, _ := New("development")
	hook := &TestHook{}
	logger.AddHook(hook)

	assert.Panics(t, func() {
		logger.Panic("panic message", map[string]any{"key": "value"})
	})

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.PanicLevel, hook.Entries[0].Level)
	assert.Equal(t, "panic message", hook.Entries[0].Message)
	assert.Equal(t, "value", hook.Entries[0].Data["key"])
}

func TestLogrusProviderFatal(t *testing.T) {
	logger, _ := New("development")
	hook := &TestHook{}
	logger.AddHook(hook)

	// Note: Testing Fatal is tricky because it exits the program.
	// You can use os.Exit with mocking if required or skip.
	t.Skip("Fatal cannot be tested without exiting the program.")
}

func TestLogrusProviderSync(t *testing.T) {
	logger, _ := New("development")
	err := logger.Sync()
	assert.NoError(t, err, "Sync should not return an error for LogrusProvider")
}
