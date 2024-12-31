package logger

import (
	"log"
	"os"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/logger"
)

func init() {
	cfg := config.GetAppConfig()
	logger.DefaultLogger = logger.New(NewDefaultLogger(cfg.Env))
	logger.DefaultLogger.Info("Default logger initiated")
}

// DefaultLogger is a custom logger that mimics LogrusProvider's methods.
type DefaultLogger struct {
	defaultFields map[string]any
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
}

// NewDefaultLogger creates a new DefaultLogger.
func NewDefaultLogger(mode string) *DefaultLogger {
	infoLogger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger := log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	logger := &DefaultLogger{
		defaultFields: make(map[string]any),
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		debugLogger:   nil,
	}

	// Enable debug logs only in non-production mode
	if mode != "production" {
		logger.debugLogger = debugLogger
		logger.Info("Logger initialized in debug mode", map[string]any{"mode": mode})
	} else {
		logger.debugLogger = log.New(nil, "", 0) // Disable debug logs
	}

	return logger
}

// WithContext creates a new DefaultLogger with contextual fields.
func (l *DefaultLogger) WithContext(fields map[string]any) *DefaultLogger {
	newLogger := *l
	newLogger.defaultFields = mergeFields(l.defaultFields, fields)
	return &newLogger
}

// Helper method to merge fields.
func mergeFields(defaultFields, fields map[string]any) map[string]any {
	merged := make(map[string]any)
	for k, v := range defaultFields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return merged
}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(msg string, fields map[string]any) {
	if l.debugLogger != nil {
		allFields := mergeFields(l.defaultFields, fields)
		l.debugLogger.Printf("%s | %v", msg, allFields)
	}
}

// Info logs an informational message.
func (l *DefaultLogger) Info(msg string, fields map[string]any) {
	allFields := mergeFields(l.defaultFields, fields)
	l.infoLogger.Printf("%s | %v", msg, allFields)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(msg string, fields map[string]any) {
	allFields := mergeFields(l.defaultFields, fields)
	l.infoLogger.Printf("WARN: %s | %v", msg, allFields)
}

// Error logs an error message.
func (l *DefaultLogger) Error(msg string, fields map[string]any) {
	allFields := mergeFields(l.defaultFields, fields)
	l.errorLogger.Printf("%s | %v", msg, allFields)
}

// DPanic logs a debug panic message and panics.
func (l *DefaultLogger) DPanic(msg string, fields map[string]any) {
	l.Error(msg, fields)
	panic(msg)
}

// Panic logs a panic message and panics.
func (l *DefaultLogger) Panic(msg string, fields map[string]any) {
	l.Error(msg, fields)
	panic(msg)
}

// Fatal logs a fatal message and exits.
func (l *DefaultLogger) Fatal(msg string, fields map[string]any) {
	allFields := mergeFields(l.defaultFields, fields)
	l.errorLogger.Printf("FATAL: %s | %v", msg, allFields)
	os.Exit(1)
}

// Sync is a no-op for DefaultLogger (similar to logrus behavior).
func (l *DefaultLogger) Sync() error {
	// Standard log doesn't need synchronization.
	return nil
}
