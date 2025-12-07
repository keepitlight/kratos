package log

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const messageKey = "message"

// Zap 日志实现 Kratos 日志记录器接口
type Zap struct {
	log  *zap.Logger
	Sync func() error
}

func (l *Zap) Log(level log.Level, kv ...interface{}) error {
	if len(kv) == 0 || len(kv)%2 != 0 {
		l.log.Warn("key value must appear in pairs: ", zap.Any("kv", kv))
		return nil
	}

	var data []zap.Field
	for i := 0; i < len(kv); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(kv[i]), kv[i+1]))
	}

	l.log.Log(zapcore.Level(level), "", data...)
	return nil
}

// NewZap 创建 Zap 日志工具
func NewZap(name, file string, level zapcore.Level, maxAge, maxSize, maxBackups int) *Zap {
	var ws zapcore.WriteSyncer

	if level == zapcore.DebugLevel {
		ws = zapcore.AddSync(os.Stdout)
	} else {
		ws = zapcore.AddSync(&lumberjack.Logger{
			Filename:   file,
			MaxAge:     maxAge,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
		})
	}

	core := zapcore.NewCore(getEncoder(), ws, level)
	opt := zap.Fields(zap.String("app", name)) // 服务名
	return &Zap{log: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2), opt), Sync: core.Sync}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.MessageKey = messageKey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}
