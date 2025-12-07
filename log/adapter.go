package log

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	// 假设您使用的是 Uber Zap
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ----------------------------------------------------
// 1. 标准库 log 包适配到 slog
// ----------------------------------------------------

// RedirectToSLog 将标准库 log 的输出重定向到指定的 slog.Logger。
func RedirectToSLog(logger *slog.Logger, level slog.Level) {
	// NewLogLogger 创建一个 log.Logger，其输出将路由到 logger 的 INFO 级别。
	// 这里我们使用 INFO 级别，因为它最符合标准 log 的预期。
	slogLogger := slog.NewLogLogger(logger.Handler(), level)

	// 设置标准库 log 包的输出目标为这个特殊的 log.Logger 实例
	log.SetOutput(slogLogger.Writer())

	// 设置 log 的 Flag，通常将时间戳和文件信息关闭，因为 slog 会提供更精确的 time/source 字段
	log.SetFlags(0)

	fmt.Println("Standard 'log' package output is now routed to slog.")
}

// ----------------------------------------------------
// 2. zap 包适配到 slog
// ----------------------------------------------------

// SlogCore 是 zapcore.Core 的实现，将 zap 的调用转发给 slog.Logger
type SlogCore struct {
	slogLogger *slog.Logger
	level      zapcore.LevelEnabler
}

// NewSlogCore 创建 SlogCore 实例
func NewSlogCore(logger *slog.Logger) zapcore.Core {
	return &SlogCore{
		slogLogger: logger,
		// 使用 zap.LevelEnablerFunc 适配 zap 的级别检查
		level: zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			// 将 zap level 转换为 slog level 进行检查
			return logger.Enabled(context.Background(), zapToSlogLevel(lvl))
		}),
	}
}

// zapToSlogLevel 将 zap 的日志级别转换为 slog 的日志级别
func zapToSlogLevel(zl zapcore.Level) slog.Level {
	switch zl {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	case zapcore.ErrorLevel:
		return slog.LevelError
	case zapcore.DPanicLevel, zapcore.FatalLevel, zapcore.PanicLevel:
		// 将致命级别映射到 Error
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Enabled 实现了 zapcore.Core 接口，检查级别是否启用
func (c *SlogCore) Enabled(lvl zapcore.Level) bool {
	return c.level.Enabled(lvl)
}

// With 实现了 zapcore.Core 接口，创建带有字段的新 Core (slog 的 With 方法)
func (c *SlogCore) With(fields []zapcore.Field) zapcore.Core {
	attrs := make([]slog.Attr, 0, len(fields))
	for _, f := range fields {
		// 将 zap 字段转换为 slog.Attr
		attrs = append(attrs, slog.Any(f.Key, f.Interface))
	}

	// 修复：slog.Logger.With 接收 variadic ...any，需要将 []slog.Attr 转换为 []any
	anyArgs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyArgs[i] = attr
	}

	// 返回一个新的 Core，包装了带有预设字段的 slog.Logger
	return NewSlogCore(c.slogLogger.With(anyArgs...))
}

// Check 实现了 zapcore.Core 接口，用于性能优化
func (c *SlogCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write 实现了 zapcore.Core 接口，执行日志写入
func (c *SlogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	slogLevel := zapToSlogLevel(entry.Level)

	// 1. 收集 zap 字段 (fields) 和 entry 字段
	attrs := make([]slog.Attr, 0, len(fields)+5)

	// 添加 zap fields
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Interface))
	}

	// 2. 使用 slog.Logger.LogAttrs 进行记录
	// 注意：这里需要处理 zap 的消息和 slog 的消息。
	// LogAttrs 接受 ...slog.Attr，因此这里可以直接解包 attrs
	c.slogLogger.LogAttrs(
		context.Background(),
		slogLevel,
		entry.Message, // zap 的消息作为 slog 的主消息
		attrs...,
	)

	// 致命错误处理 (例如 zap.Fatal/Panic)
	if entry.Level >= zapcore.FatalLevel {
		os.Exit(1)
	}
	return nil
}

// Sync 实现了 zapcore.Core 接口，同步写入
func (c *SlogCore) Sync() error {
	// slog 的 handler 负责同步，通常只需刷新输出流，这里不做特殊处理
	return nil
}
