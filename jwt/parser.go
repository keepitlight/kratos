package jwt

import (
	"context"
	"strings"

	"github.com/keepitlight/kratos/net/grpc"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	j5 "github.com/golang-jwt/jwt/v5"
)

type Parser struct {
	Name string // 解析器名称

	secretKey     []byte           // 签署密钥，对于非对称加密的算法，这个字段为公钥
	signingMethod j5.SigningMethod // 签署方法/算法
}

// NewParser to create a JWT parser, parameters signingMethod is the signing method/algorithm,
// parameters secretKey is the signing key
//
// 创建一个 JWT 解析器，参数 signingMethod 是签署方法/算法，参数 secretKey 是签署密钥
func NewParser(signingMethod j5.SigningMethod, secretKey []byte) *Parser {
	j5.NewParser()
	return &Parser{
		secretKey:     secretKey,
		signingMethod: signingMethod,
	}
}

// Parse parses the JWT string and returns a Claims object.
//
// 解析 JWT，返回 Claims
func (p *Parser) Parse(jwt string) (claims *Claims, err error) {
	var opts []j5.ParserOption

	if p.Name != "" {
		opts = append(opts, j5.WithAudience(p.Name))
	}

	token, err := j5.ParseWithClaims(
		jwt,
		&Claims{},
		func(token *j5.Token) (interface{}, error) {
			if strings.Compare(token.Method.Alg(), p.signingMethod.Alg()) != 0 {
				return nil, j5.ErrSignatureInvalid
			}
			return p.secretKey, nil
		},
		opts...,
	)

	if err != nil {
		return
	}
	var ok bool
	if claims, ok = token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, j5.ErrSignatureInvalid
}

// Lookup to get the JWT from the current HTTP request context or gRPC metadata,
// and parse it, return nil if not found otherwise return Claims object
//
// 在当前 HTTP 请求的上下文中获取认证头，或者从 gRPC 的 metadata 中获取认证头，并解析为 JWT，
// 如果未找到则返回 nil，否则返回 Claims 对象
func (p *Parser) Lookup(ctx context.Context) (claims *Claims, err error) {
	// 获取请求头
	if tr, yes := transport.FromServerContext(ctx); yes && tr.Kind() == transport.KindHTTP {
		if ht, yes := tr.(*http.Transport); yes {
			a := ht.RequestHeader().Get(Authorization)
			if strings.HasPrefix(a, grpc.BearerPrefix) {
				t, e := p.Parse(a[7:])
				return t, e
			}
		}
	} else {
		if v, yes := grpc.LookupToken(ctx); yes {
			t, e := p.Parse(v)
			return t, e
		}
	}
	return nil, nil
}
