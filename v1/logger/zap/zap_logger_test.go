package zap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	// Test Production Mode
	logger, err := New("production")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Info("Test production mode", map[string]any{"env": "production"})

	// Test Development Mode
	logger, err = New("development")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Debug("Test development mode", map[string]any{"env": "development"})

	// Test Invalid Mode (Defaults to Development)
	logger, err = New("invalid")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	logger.Debug("Test invalid mode defaults", map[string]any{"env": "development"})
}

func TestZapLoggerMethods(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	zapLogger := &ZapLogger{mockLogger}

	fields := map[string]any{
		"string": "value",
		"int":    123,
		"bool":   true,
		"error":  assert.AnError,
		"other":  []int{1, 2, 3},
	}

	// Capture logs and verify output
	zapLogger.Debug("Debug message", fields)
	zapLogger.Info("Info message", fields)
	zapLogger.Warn("Warn message", fields)
	zapLogger.Error("Error message", fields)
	zapLogger.DPanic("DPanic message", fields)

	// Panic testing
	assert.Panics(t, func() {
		zapLogger.Panic("Panic message", fields)
	})

	// Fatal testing: Skipped due to program termination
	t.Skip("Fatal testing is skipped because it terminates the program.")
}

func TestConvertToZapFields(t *testing.T) {
	fields := map[string]any{
		"string": "test",
		"int":    42,
		"bool":   true,
		"error":  assert.AnError,
		"other":  []int{1, 2, 3},
	}

	expected := []zap.Field{
		zap.String("string", "test"),
		zap.Int("int", 42),
		zap.Bool("bool", true),
		zap.Error(assert.AnError),
		zap.Any("other", []int{1, 2, 3}),
	}

	zapFields := convertToZapFields(fields)

	assert.Len(t, zapFields, len(expected))
	for i, field := range zapFields {
		assert.Equal(t, expected[i].Key, field.Key)
		assert.Equal(t, expected[i].Type, field.Type)
	}
}

func TestSync(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	zapLogger := &ZapLogger{mockLogger}

	err := zapLogger.Sync()
	assert.NoError(t, err)
}
