package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init 初始化日志
func Init(mode string) error {
	var cfg zap.Config

	if mode == "release" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.OutputPaths = []string{"stdout", "logs/starnet.log"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	var err error
	Log, err = cfg.Build()
	if err != nil {
		return err
	}

	Log.Info("Logger initialized", zap.String("mode", mode))
	return nil
}

// Sync 刷新日志缓冲
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
