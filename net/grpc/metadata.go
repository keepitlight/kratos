package grpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

const (
	Authorization = "Authorization"
)

// LookupToken 从 context 中查找 gRPC 元数据中注入的 JWT
func LookupToken(ctx context.Context) (jwt string, ok bool) {
	if tr, yes := transport.FromServerContext(ctx); yes && tr.Kind() == transport.KindGRPC {
		if md, yes := metadata.FromIncomingContext(ctx); yes {
			auth := md.Get(Authorization)
			if len(auth) <= 7 {
				return
			}
			// 获取 token
			jwt = auth[0][7:]
			ok = true
		} else {
			return
		}
		// 获取 Authorization header
	}
	return
}

// InjectToken 通过 context 将 JWT 注入到 gRPC 元数据中
func InjectToken(ctx context.Context, jwt string) context.Context {
	if tr, yes := transport.FromServerContext(ctx); yes && tr.Kind() == transport.KindGRPC {
		// 在 metadata 中添加 Authorization header
		md := metadata.Pairs(Authorization, "Bearer "+jwt)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
