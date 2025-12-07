package log

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.uber.org/zap"
)

func ExampleSetupStandardLogRedirect() {
	// I. 统一配置：创建一个最终的 slog Handler
	// 所有日志都将通过这个 Handler 输出为 JSON 格式
	slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// 使用我们之前定义的配置函数来处理脱敏/重命名
		// ReplaceAttr: logconfig.NewReplaceAttrFunc(logconfig.DefaultConfig()),
	})
	unifiedSlogLogger := slog.New(slogHandler)

	fmt.Println("---------------------------------")
	fmt.Println("1. 统一标准库 log")
	fmt.Println("---------------------------------")

	// II. 适配标准库 log
	RedirectToSLog(unifiedSlogLogger, slog.LevelInfo)
	log.Println("This is a legacy log message.")

	// III. 适配 zap
	fmt.Println("\n---------------------------------")
	fmt.Println("2. 统一 Uber Zap")
	fmt.Println("---------------------------------")

	// 1. 创建 zap.Logger
	slogCore := NewSlogCore(unifiedSlogLogger)
	zapLogger := zap.New(slogCore)
	defer func(zapLogger *zap.Logger) { _ = zapLogger.Sync() }(zapLogger) // 确保所有日志被写入

	// 2. 使用 zap.Logger 记录日志
	zapLogger.Info("Zap info message with context",
		zap.String("component", "database"),
		zap.Int("query_ms", 50),
	)

	// IV. 直接使用 slog (确保适配器工作正常时，可以直接使用统一的 logger)
	fmt.Println("\n---------------------------------")
	fmt.Println("3. 直接使用 slog")
	fmt.Println("---------------------------------")
	unifiedSlogLogger.Debug("Slog debug message (should appear).", "worker_id", 7)

	// output
	// ---------------------------------
	// 1. 统一标准库 log
	// ---------------------------------
	// 2023/08/05 17:05:07 This is a legacy log message.
	// ---------------------------------
	// 2. 统一 Uber Zap
	// ---------------------------------
	// {"time":"2023-08-05T17:05:07.000+08:00","level":"info","msg":"Zap info message with context","component":"database","query_ms":50}
	// ---------------------------------
	// 3. 直接使用 slog
	// ---------------------------------
	// {"time":"2023-08-05T17:05:07.000+08:00","level":"debug","msg":"Slog debug message (should appear).","worker_id":7}

}
