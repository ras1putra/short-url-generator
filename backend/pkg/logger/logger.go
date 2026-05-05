package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

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
