package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type loggerCtxKey struct{}

func Init(env string) error {
	cores := []zapcore.Core{
		zapcore.NewCore(
			getEncoder(env),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		),
	}

	if env == "development" {
		logWriter := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		cores = append(cores, zapcore.NewCore(
			getEncoder(env),
			zapcore.AddSync(logWriter),
			zap.InfoLevel,
		))
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)

	return nil
}

func getEncoder(env string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if env == "development" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func WithUser(userID interface{}) *zap.Logger {
	return zap.L().With(zap.Any("user_id", userID))
}

func WithFields(fields ...zap.Field) *zap.Logger {
	return zap.L().With(fields...)
}

func Ctx(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(loggerCtxKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.L()
}

func WithCtx(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, log)
}
