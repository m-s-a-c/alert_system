package logging

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func InitLogging(mode, logDir string) {
	logPath := filepath.Join(logDir, "log/alert.log")
	logWriter := getWriterSyncer(logPath)

	cfg := getZapConfig(mode)
	cfg.Level.UnmarshalText([]byte(viper.GetString("logging.level")))

	if mode == "development" {
		logWriter = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), logWriter)
	}

	logger, err := cfg.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(getEncoder(cfg.Encoding, cfg.EncoderConfig), logWriter, cfg.Level)
	}))
	if err != nil {
		panic(err)
	}

	Logger = logger
}

func getZapConfig(mode string) zap.Config {
	cfg := zap.NewProductionConfig()

	if mode == "development" {
		cfg := zap.NewDevelopmentConfig()
		cfg.DisableCaller = false
	}

	cfg.Encoding = "console"
	cfg.EncoderConfig = getEncoderConfig()

	return cfg
}

func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "name",
		MessageKey:    "msg",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
	}
}

func getEncoder(encoding string, config zapcore.EncoderConfig) zapcore.Encoder {
	switch encoding {
	case "json":
		return zapcore.NewJSONEncoder(config)
	case "console":
		return zapcore.NewConsoleEncoder(config)
	default:
		zap.L().Fatal("Invalid logging Encoding", zap.String("encoding", encoding))
		return nil
	}
}

func getWriterSyncer(logPath string) zapcore.WriteSyncer {
	var ioWriter = &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, //MB
		MaxBackups: 3,  //number of backups
		MaxAge:     28, //days
		LocalTime:  true,
		Compress:   false, //disabled by default
	}
	ioWriter.Rotate()
	return zapcore.AddSync(ioWriter)
}
