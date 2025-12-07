package log

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/keepitlight/kratos/runtime"
)

// T 为组合的 slog.Handler
type T struct {
	handlers []slog.Handler
}

// Join 是 T 的构造函数，用于组合日志处理器
func Join(handlers ...slog.Handler) *T {
	return &T{
		handlers: handlers,
	}
}

func Console(level slog.Level, addSource bool) slog.Handler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	})
}

func JSON(level slog.Level, file string, addSource bool) slog.Handler {
	return File(level, file, true, addSource)
}

// File 创建一个日志文件处理器，将控制台和文件作为内部 Handler
func File(level slog.Level, file string, json, addSource bool) slog.Handler {
	logFile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("无法打开日志文件", "error", err)
		return nil
	}

	runtime.Defer(func(logger log.Logger) {
		e := logFile.Close()
		if e != nil {
			return
		}
	})

	if json {
		return slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level:     level,
			AddSource: addSource,
		})
	}
	return slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	})
}

// Enabled 方法: 告诉 Logger 是否需要执行 Handle。
// 只要有一个内部 Handler 对该级别感兴趣，就返回 true。
func (h *T) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler != nil && handler.Enabled(ctx, level) {
			return true
		}
	}
	return false // 所有 Handler 都不启用
}

// Handle 方法: 核心路由逻辑。根据每个 Handler 自身的 Level 配置决定是否处理 Record。
func (h *T) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		// 必须先检查每个 Handler 是否启用，因为它们有不同的 Level 配置
		if handler != nil && handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				// 通常的做法是记录错误，但不中断后续 Handler 的执行
				// 为了简洁，这里仅打印错误并继续
				// fmt.Fprintf(os.Stderr, "MultiHandler: error handling log: %v\n", err)
			}
		}
	}
	return nil
}

// WithAttrs 方法: 为两个内部 Handler 添加共享的属性
func (h *T) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		if handler != nil {
			newHandlers[i] = handler.WithAttrs(attrs)
		}
	}
	return &T{handlers: newHandlers}
}

// WithGroup 方法: 为两个内部 Handler 添加共享的分组
func (h *T) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		if handler != nil {
			newHandlers[i] = handler.WithGroup(name)
		}
	}
	return &T{handlers: newHandlers}
}

type kratosLoggerAdapter struct {
	logger *slog.Logger
	msgKey string
}

func (k *kratosLoggerAdapter) mapLevel(level log.Level) slog.Level {
	switch level {
	case log.LevelDebug:
		return slog.LevelDebug
	case log.LevelWarn:
		return slog.LevelWarn
	case log.LevelError, log.LevelFatal:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (k *kratosLoggerAdapter) Log(level log.Level, kv ...any) error {
	l := k.mapLevel(level)
	if !k.logger.Enabled(context.Background(), l) {
		return nil
	}

	// 1. 提取 msg 和剩余的属性
	msg := "<nil>" // 默认消息
	var attrs []slog.Attr

	for i := 0; i < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			// 如果 key 不是字符串，跳过这一对，或者记录一个警告
			continue
		}

		val := kv[i+1] // 假设 kv 总是偶数个

		// 检查这个键是否是我们定义的消息键
		if strings.EqualFold(key, k.msgKey) {
			if vStr, isStr := val.(string); isStr {
				msg = vStr
			}
			// 不将 msg 字段再次作为普通属性添加
			continue
		}

		// 否则，作为普通属性添加
		attrs = append(attrs, slog.Any(key, val))
	}

	// 2. 将所有的键值对（除了 msg 键）作为 Attr 传入
	k.logger.LogAttrs(context.Background(), l, msg, attrs...)
	return nil
}

func (h *T) Logger(msgKey string) log.Logger {
	if len(msgKey) == 0 {
		msgKey = log.DefaultMessageKey
	}
	return &kratosLoggerAdapter{logger: slog.New(h), msgKey: msgKey}
}
func (h *T) DefaultLogger() log.Logger {
	return h.Logger(log.DefaultMessageKey)
}

// Cast 将 slog.Logger 转为 kratos Logger
func Cast(logger *slog.Logger, msgKey string) log.Logger {
	key := log.DefaultMessageKey
	if len(msgKey) > 0 {
		key = msgKey
	}
	return &kratosLoggerAdapter{
		logger: logger,
		msgKey: key,
	}
}

// Default 将默认的 slog Logger 转为 kratos Logger 返回
func Default() log.Logger {
	return &kratosLoggerAdapter{
		logger: slog.Default(),
		msgKey: log.DefaultMessageKey,
	}
}
