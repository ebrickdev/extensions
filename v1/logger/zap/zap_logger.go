package zap

import (
	"fmt"
	"log"
	"time"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/logger"
	"go.uber.org/zap"
)

func Init() logger.Logger {
	cfg := config.GetAppConfig()
	l, err := (New(cfg.Env))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	return logger.New(l)
}

// ZapLogger is a concrete implementation of the Logger interface using zap.Logger.
type ZapLogger struct {
	*zap.Logger
}

// New initializes a new zap-based Logger.
// It accepts the environment as a parameter to configure the logger appropriately.
func New(mode string) (*ZapLogger, error) {
	var zapConfig zap.Config

	switch mode {
	case "production":
		zapConfig = zap.NewProductionConfig()
	default:
		zapConfig = zap.NewDevelopmentConfig()
		fmt.Printf("Warning: invalid mode '%s', defaulting to development mode\n", mode)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize zap logger in %s mode: %w", mode, err)
	}

	// Set default fields
	logger.Info("Logger initialized", zap.String("mode", mode))
	return &ZapLogger{logger}, nil
}

// convertToZapFields converts custom Field instances to zap.Field with type-specific handling.
func convertToZapFields(fields map[string]any) []zap.Field {
	if fields == nil {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		case time.Time:
			zapFields = append(zapFields, zap.Time(key, v))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}
	return zapFields
}

// WithContext creates a new ZapLogger with contextual fields.
func (z *ZapLogger) WithContext(fields map[string]any) *ZapLogger {
	return &ZapLogger{z.Logger.With(convertToZapFields(fields)...)}
}

// Implementing the Logger interface methods for ZapLogger.
func (z *ZapLogger) Debug(msg string, fields map[string]any) {
	z.Logger.Debug(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields map[string]any) {
	z.Logger.Info(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields map[string]any) {
	z.Logger.Warn(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields map[string]any) {
	z.Logger.Error(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) DPanic(msg string, fields map[string]any) {
	z.Logger.DPanic(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Panic(msg string, fields map[string]any) {
	z.Logger.Panic(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Fatal(msg string, fields map[string]any) {
	z.Logger.Fatal(msg, convertToZapFields(fields)...)
}

func (z *ZapLogger) Sync() error {
	return z.Logger.Sync()
}
