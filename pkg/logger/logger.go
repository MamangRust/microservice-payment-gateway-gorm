package logger

import (
	"os"

	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerInterface interface {
	Info(message string, fields ...zap.Field)
	Fatal(message string, fields ...zap.Field)
	Debug(message string, fields ...zap.Field)
	Error(message string, fields ...zap.Field)
	Warn(message string, fields ...zap.Field)
	Check(level zapcore.Level, message string) *zapcore.CheckedEntry
	With(fields ...zap.Field) LoggerInterface
	Sync() error
}

type Logger struct {
	Log *zap.Logger
}

var instance LoggerInterface

func NewLogger(service string, loggerProvider *log.LoggerProvider) (LoggerInterface, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	stdoutCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// Temporary: bypass otelzap to avoid systemic protobuf panic
	// core := zapcore.NewTee(stdoutCore, otelCore)
	core := stdoutCore

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.FatalLevel),
	).With(zap.String("service", service))

	l := &Logger{Log: logger}
	instance = l
	return l, nil
}

func (l *Logger) Info(message string, fields ...zap.Field) {
	l.Log.Info(message, fields...)
}

func (l *Logger) Fatal(message string, fields ...zap.Field) {
	l.Log.Fatal(message, fields...)
}

func (l *Logger) Debug(message string, fields ...zap.Field) {
	l.Log.Debug(message, fields...)
}

func (l *Logger) Error(message string, fields ...zap.Field) {
	l.Log.Error(message, fields...)
}

func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.Log.Warn(message, fields...)
}

func (l *Logger) Check(level zapcore.Level, message string) *zapcore.CheckedEntry {
	return l.Log.Check(level, message)
}

func (l *Logger) With(fields ...zap.Field) LoggerInterface {
	return &Logger{Log: l.Log.With(fields...)}
}

func (l *Logger) Sync() error {
	return l.Log.Sync()
}

func GetInstance() LoggerInterface {
	return instance
}

func ResetInstance() {
	instance = nil
}

// NoopLogger discards every log call. It is used by components (e.g. the
// dependency guard) that require a LoggerInterface but should not emit logs,
// and by unit tests that do not want log noise.
type NoopLogger struct{}

func (NoopLogger) Info(string, ...zap.Field)       {}
func (NoopLogger) Fatal(string, ...zap.Field)      {}
func (NoopLogger) Debug(string, ...zap.Field)      {}
func (NoopLogger) Error(string, ...zap.Field)      {}
func (NoopLogger) Warn(string, ...zap.Field)       {}
func (NoopLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry { return nil }
func (NoopLogger) With(...zap.Field) LoggerInterface { return NoopLogger{} }
func (NoopLogger) Sync() error { return nil }
