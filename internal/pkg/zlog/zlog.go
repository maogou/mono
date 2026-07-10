package zlog

import (
	"context"
	"os"
	"sync"
	"time"

	"go_template/internal/config"
	"go_template/internal/constant"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *Logger
	once sync.Once
)

type contextKey string

const CtxLoggerKey contextKey = "zapLogger"

type Logger struct {
	*zap.Logger
}

func NewZapLog(conf *config.Config) *Logger {
	once.Do(func() {
		log = buildLog(conf)
	})
	return log
}

func buildLog(conf *config.Config) *Logger {
	hook := &lumberjack.Logger{
		Filename:   lo.Ternary(len(conf.Log.Filename) == 0, "go_template.log", conf.Log.Filename),
		MaxSize:    conf.Log.MaxSize,
		MaxAge:     conf.Log.MaxAge,
		MaxBackups: conf.Log.MaxBackups,
		Compress:   conf.Log.Compress,
	}

	var encoder zapcore.Encoder

	if conf.Log.Encoding != constant.Console {
		encoder = zapcore.NewJSONEncoder(
			zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     timeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		)
	} else {
		encoder = zapcore.NewConsoleEncoder(
			zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "Logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
				EncodeTime:     timeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.FullCallerEncoder,
			},
		)
	}

	var level = levelFromString(conf.Log.Level)
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)),
		level,
	)

	if conf.Log.Mode != constant.Console {
		if conf.Log.Mode == constant.File {
			core = zapcore.NewCore(encoder, zapcore.AddSync(hook), level)
		}
	} else {
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	}

	if conf.Mode == constant.Release {
		return &Logger{zap.New(
			core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel),
		)}
	}
	return &Logger{zap.New(
		core, zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel),
	)}
}

func levelFromString(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func (l *Logger) V(ctx context.Context, fields ...zapcore.Field) context.Context {
	if c, ok := ctx.(*gin.Context); ok {
		ctx = c.Request.Context()
		c.Request = c.Request.WithContext(
			context.WithValue(
				ctx,
				CtxLoggerKey,
				l.C(ctx).With(fields...),
			),
		)
		return c
	}
	return context.WithValue(ctx, CtxLoggerKey, l.C(ctx).With(fields...))
}

func (l *Logger) C(ctx context.Context) *Logger {
	if c, ok := ctx.(*gin.Context); ok {
		ctx = c.Request.Context()
	}
	zl := ctx.Value(CtxLoggerKey)
	ctxLogger, ok := zl.(*zap.Logger)
	if ok {
		return &Logger{ctxLogger}
	}
	return l
}

func defaultLog() *Logger {
	once.Do(func() {
		log = buildDefaultLog()
	})
	return log
}

func buildDefaultLog() *Logger {
	encoder := zapcore.NewJSONEncoder(
		zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	)

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)

	return &Logger{zap.New(
		core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel),
	)}
}

func C(ctx context.Context) *Logger {
	if c, ok := ctx.(*gin.Context); ok {
		ctx = c.Request.Context()
	}
	zl := ctx.Value(CtxLoggerKey)
	if ctxLogger, ok := zl.(*zap.Logger); ok && ctxLogger != nil {
		return &Logger{ctxLogger}
	}

	if log != nil {
		return log
	}

	return defaultLog()
}
