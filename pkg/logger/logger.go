package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LoggerKey   = "logger"
	RequestID   = "requestID"
	ServiceName = "service"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
}

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func (l logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Debug(msg, fields...)
}

func (l logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Info(msg, fields...)
}

func (l logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Warn(msg, fields...)
}

func (l logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Error(msg, fields...)
}

func (l logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Fatal(msg, fields...)
}

func New(level zapcore.Level, serviceName string) Logger {
	config := zap.Config{
		Encoding:         "console", // Используем консольный вывод
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Цветной вывод уровня лога
			EncodeTime:     zapcore.ISO8601TimeEncoder,       // Читаемое отображение времени
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
	}
	zapLogger, _ := config.Build()
	defer zapLogger.Sync()

	return &logger{
		serviceName: serviceName,
		logger:      zapLogger,
	}
}

func GetLoggerFromCtx(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}
