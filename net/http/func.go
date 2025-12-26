package http

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Only apply a middleware to all HTTP requests
//
// HTTP 专用的中间件应用到所有 HTTP 请求
func Only(handler func(ctx context.Context, req interface{})) middleware.Middleware {
	return func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 检查是否为 HTTP 请求
			if tr, ok := transport.FromServerContext(ctx); ok && tr.Kind() == transport.KindHTTP {
				// 只有 HTTP 请求才执行这里的逻辑
				handler(ctx, req)
			}

			return h(ctx, req)
		}
	}
}

// IsHTTPRequest determines whether the context is an HTTP request.
//
// 检查上下文是否为 HTTP 请求
func IsHTTPRequest(ctx context.Context) bool {
	// 检查是否为 HTTP 请求
	if tr, ok := transport.FromServerContext(ctx); ok && tr.Kind() == transport.KindHTTP {
		return true
	}
	return false
}

// SetHeader to set response headers, argument overwrite is whether to overwrite the existing header
//
// 设置响应头，参数 overwrite 指示是否覆盖已有的头
func SetHeader(ctx context.Context, key, value string, overwrite bool) bool {
	// 判断是否为 HTTP 请求并设置响应头
	if tr, ok := transport.FromServerContext(ctx); ok && tr.Kind() == transport.KindHTTP {
		if ht, ok := tr.(*http.Transport); ok {
			// 设置响应头
			if overwrite {
				ht.ReplyHeader().Set(key, value)
			} else {
				ht.ReplyHeader().Add(key, value)
			}
			return true
		}
	}
	return false
}

// AddHeader to add a response header and not overwrite existing headers.
//
// 添加响应头，不覆盖已存在的头
func AddHeader(ctx context.Context, key, value string) bool {
	return SetHeader(ctx, key, value, false)
}

// LookupHeader to get a request header.
//
// 获取指定请求头的所的值
func LookupHeader(ctx context.Context, key string) (values []string, ok bool) {
	// 获取请求头
	if tr, yes := transport.FromServerContext(ctx); yes && tr.Kind() == transport.KindHTTP {
		if ht, yes := tr.(*http.Transport); yes {
			return ht.RequestHeader().Values(key), true
		}
	}
	return nil, false
}
