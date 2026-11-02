package interceptors

import (
	"context"
	"fmt"

	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
)

func LoggerInterceptor(l *logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		zapFields := make([]zap.Field, 0, (len(fields)+1)/2)

		// Convert key-value pairs into zap.Field.
		for i := 0; i+1 < len(fields); i += 2 {
			key := fmt.Sprint(fields[i])
			value := fields[i+1]

			zapFields = append(zapFields, zap.Any(key, value))
		}

		// Handle an unexpected unpaired field.
		if len(fields)%2 != 0 {
			zapFields = append(
				zapFields,
				zap.Any("unpaired_field", fields[len(fields)-1]),
			)
		}

		log := l.WithOptions(zap.AddCallerSkip(1)).With(zapFields...)

		switch lvl {
		case logging.LevelDebug:
			log.Debug(msg)
		case logging.LevelInfo:
			log.Info(msg)
		case logging.LevelWarn:
			log.Warn(msg)
		case logging.LevelError:
			log.Error(msg)
		default:
			log.Warn(msg, zap.Any("unknown_log_level", lvl))
		}
	})
}
