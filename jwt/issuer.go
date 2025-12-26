package jwt

import (
	"time"

	j5 "github.com/golang-jwt/jwt/v5"
)

// Issuer defines a JWT issuer.
//
// JWT 签发者
type Issuer struct {
	name       string   // 签署人名称
	audiences  []string // 受众列表
	signingKey string   // 签署密钥，对于非对称加密的算法，这个字段为私钥
	// parseKey      string        // 解析密钥，对于非对称加密的算法，这个字段为公钥
	signingMethod string        // 签署方法/算法
	ttl           time.Duration // 令牌有效期
}

// NewIssuer 创建一个 JWT 签发者，参数 signingMethod 签署方法/算法，参数 signingKey 签署密钥，
// 参数 name 签发者名称，参数 audiences 消费端列表
func NewIssuer(signingMethod, signingKey, name string, ttl time.Duration, audiences ...string) *Issuer {
	return &Issuer{
		name:          name,
		audiences:     audiences,
		signingKey:    signingKey,
		signingMethod: signingMethod,
		ttl:           ttl,
	}
}

// RefreshTokenIssuer 创建一个使用对称加密算法 HS256 的刷新令牌签发者，参数 secureKey 密钥，参数 ttl 令牌有效期
func RefreshTokenIssuer(secureKey string, ttl time.Duration) *Issuer {
	return NewIssuer("HS256", secureKey, "", ttl)
}

// AccessTokenIssuer 创建一个使用非对称加密算法 ES256 的访问令牌签发者，参数 privateKey 私钥（仅签发者，即认证服务持有），
// 另有 publicKey 公钥分发给所有消费服务，参数 name 签发者名称，参数 ttl 令牌有效期，参数 audiences 受众列表
func AccessTokenIssuer(privateKey, name string, ttl time.Duration, audiences ...string) *Issuer {
	return NewIssuer("ES256", privateKey, name, ttl, audiences...)
}

func (i *Issuer) Make(id, subject string, tags ...string) (claims *Claims) {
	now := time.Now()
	expiresAt := now.Add(i.ttl)

	claims = &Claims{
		RegisteredClaims: j5.RegisteredClaims{
			Issuer:    i.name,
			Subject:   subject,
			ID:        id,
			Audience:  i.audiences,
			ExpiresAt: j5.NewNumericDate(expiresAt),
			NotBefore: j5.NewNumericDate(now),
			IssuedAt:  j5.NewNumericDate(now),
		},
		Tags: tags,
	}
	return
}

// Sign 签署一个 JWT，返回签名后的字符串
func (i *Issuer) Sign(id, subject string, tags ...string) (jwt string, err error) {
	claims := i.Make(id, subject, tags...)
	token := j5.NewWithClaims(j5.GetSigningMethod(i.signingMethod), claims)
	return token.SignedString([]byte(i.signingKey))
}

func (i *Issuer) Generate(id, subject string, tags ...string) (token *Token, err error) {
	v, err := i.Sign(id, subject, tags...)
	if err != nil {
		return nil, err
	}
	return &Token{
		Value: v,
		Ttl:   i.ttl,
	}, nil
}
