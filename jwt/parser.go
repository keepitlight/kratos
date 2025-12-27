package jwt

import j5 "github.com/golang-jwt/jwt/v5"

type Parser struct {
	secretKey     []byte // 签署密钥，对于非对称加密的算法，这个字段为公钥
	signingMethod string // 签署方法/算法
}

func NewParser(secretKey []byte, signingMethod string) *Parser {
	return &Parser{
		secretKey:     secretKey,
		signingMethod: signingMethod,
	}
}

// Parse 解析 JWT，返回 Claims
func (p *Parser) Parse(jwt string) (claims *Claims, err error) {
	token, err := j5.ParseWithClaims(jwt, &Claims{}, func(token *j5.Token) (interface{}, error) {
		return p.secretKey, nil
	})
	if err != nil {
		return
	}
	var ok bool
	if claims, ok = token.Claims.(*Claims); ok && token.Valid {
		return
	}
	return nil, j5.ErrSignatureInvalid
}
