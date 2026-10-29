package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Nurlan270/task-tracker-go/common/config"
)

const (
	envDev   = "dev"
	envLocal = "local"
	envProd  = "prod"
)

type Logger struct {
	*zap.Logger

	file *os.File
}

var (
	globalLogger *Logger
	once         sync.Once
)

func Init(cfg config.Config) {
	once.Do(func() {
		env := cfg.GetAppEnv()

		encoder := getEncoder(env)

		ws, file := getWriteSyncer(env, cfg.GetAppName())

		enabler := getLevelEnabler(env)

		core := zapcore.NewCore(encoder, ws, enabler)

		zapLogger := zap.New(core, zap.AddCaller())

		globalLogger = &Logger{
			Logger: zapLogger,
			file:   file,
		}
	})
}

// Get returns global logger's instance.
//
// It will panic, if logger is not initialized yet.
func Get() *Logger {
	if globalLogger == nil || globalLogger.Logger == nil {
		panic("logger: call on nil logger instance, please call Init() first")
	}

	return globalLogger
}

// SetOp sets "op" context attribute with provided value
// which should point to operation where logger was invoked from.
func (l *Logger) SetOp(v string) *zap.Logger {
	return l.With(zap.String("op", v))
}

func (l *Logger) Close() error {
	if l == nil {
		return nil
	}

	var errs []error

	if l.Logger != nil {
		errs = append(errs, l.Sync())
		l.Logger = nil
	}

	if l.file != nil {
		errs = append(errs, l.file.Close())
		l.file = nil
	}

	return errors.Join(errs...)
}

func getEncoder(env string) zapcore.Encoder {
	encoderCfg := buildEncoderConfig(env)

	var encoder zapcore.Encoder

	switch env {
	case envDev, envLocal:
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	case envProd:
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	return encoder
}

func getWriteSyncer(env, appName string) (zapcore.WriteSyncer, *os.File) {
	var (
		ws   zapcore.WriteSyncer
		file *os.File
	)

	switch env {
	case envDev, envLocal:
		ws = zapcore.AddSync(os.Stdout)
	case envProd:
		file = buildLogFile(appName)
		ws = zapcore.Lock(zapcore.AddSync(file))
	default:
		file = buildLogFile(appName)
		ws = zapcore.Lock(zapcore.AddSync(file))
	}

	return ws, file
}

func getLevelEnabler(env string) zapcore.LevelEnabler {
	var enabler zapcore.LevelEnabler

	switch env {
	case envDev, envLocal:
		enabler = zapcore.DebugLevel
	case envProd:
		enabler = zapcore.InfoLevel
	default:
		enabler = zapcore.InfoLevel
	}

	return enabler
}

func buildEncoderConfig(env string) zapcore.EncoderConfig {
	var levelEncoder zapcore.LevelEncoder

	switch env {
	case envDev, envLocal:
		levelEncoder = zapcore.CapitalColorLevelEncoder
	case envProd:
		levelEncoder = zapcore.CapitalLevelEncoder
	default:
		levelEncoder = zapcore.CapitalLevelEncoder
	}

	return zapcore.EncoderConfig{
		NameKey:        "logger",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "message",
		TimeKey:        "timestamp",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeName:     zapcore.FullNameEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    levelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
}

func buildLogFile(appName string) *os.File {
	logFilePath := filepath.Join(
		"logs", strcase.ToSnake(appName)+".log",
	)

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		panic(fmt.Sprintf("logger: failed to create log file: %s", err))
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("logger: failed to open log file: %s", err))
	}

	return file
}
