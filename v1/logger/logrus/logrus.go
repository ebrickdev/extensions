package logrus

import (
	"log"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/logger"
	"github.com/sirupsen/logrus"
)

func init() {
	cfg := config.GetAppConfig()
	l, err := (New(cfg.Env))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logger.DefaultLogger = logger.New(l)
	logger.DefaultLogger.Info("Logrus logger initiated")
}

// LogrusProvider is a concrete implementation of the Logger interface using logrus.Logger.
type LogrusProvider struct {
	*logrus.Logger
	defaultFields map[string]any
}

// NewLogrusProvider initializes a new logrus-based Logger.
func New(mode string) (*LogrusProvider, error) {
	logger := logrus.New()

	switch mode {
	case "production":
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	default:
		logger.SetFormatter(&logrus.TextFormatter{})
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{"mode": mode}).Info("Logger initialized")

	return &LogrusProvider{
		logger,
		make(map[string]any),
	}, nil
}

// WithContext creates a new LogrusProvider with contextual fields.
func (l *LogrusProvider) WithContext(fields map[string]any) *LogrusProvider {
	return &LogrusProvider{l.Logger.WithFields(fields).Logger, fields}
}

// Implementing the Logger interface methods for LogrusProvider.
func (l *LogrusProvider) Debug(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Debug(msg)
}

func (l *LogrusProvider) Info(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Info(msg)
}

func (l *LogrusProvider) Warn(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(l.defaultFields).WithFields(fields).Warn(msg)
}

func (l *LogrusProvider) Error(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Error(msg)
}

func (l *LogrusProvider) DPanic(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Panic(msg)

	if l.Logger.GetLevel() <= logrus.DebugLevel {
		panic(msg)
	}
}

func (l *LogrusProvider) Panic(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Panic(msg)
}

func (l *LogrusProvider) Fatal(msg string, fields map[string]any) {
	l.Logger.WithFields(l.defaultFields).WithFields(fields).Fatal(msg)
}

func (l *LogrusProvider) Sync() error {
	// logrus does not require explicit synchronization.
	return nil
}
