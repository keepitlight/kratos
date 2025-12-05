package log

import (
	"context"
	"log/slog"
	"os"
)

func ExampleNew() {
	// 最低记录级别设置为 INFO，格式为 Text
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // 控制台: 仅记录 INFO 及以上级别
	})

	// 3. 组合 T
	t := New(consoleHandler)

	// 4. 创建 Logger
	logger := slog.New(t)
	slog.SetDefault(logger)

	// --- 记录测试日志 ---
	slog.Debug("配置加载完成，这是 Debug 消息")
	slog.Info("应用程序启动", "version", "1.0.0")
	slog.Warn("数据库连接响应慢", "latency_ms", 500)
	slog.Error("处理订单失败", "order_id", 999, "err", context.DeadlineExceeded)

	logger.Info("测试结束，请查看控制台和 app_debug.log 文件")

	// output:
	// time=2025-12-05T21:29:43.155+08:00 level=INFO msg=应用程序启动 version=1.0.0
	// time=2025-12-05T21:29:43.156+08:00 level=WARN msg=数据库连接响应慢 latency_ms=500
	// time=2025-12-05T21:29:43.156+08:00 level=ERROR msg=处理订单失败 order_id=999 err="context deadline exceeded"
	// time=2025-12-05T21:29:43.156+08:00 level=INFO msg="测试结束，请查看控制台和 app_debug.log 文件"

}
