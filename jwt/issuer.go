package jwt

import (
	"time"

	j5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	DefaultIssuerName      = "AuthManager"         // 默认签发者名称
	DefaultSigningMethod   = j5.SigningMethodHS256 // 默认的对称加密签署方法/算法
	DefaultPKSigningMethod = j5.SigningMethodES256 // 默认的非对称加密签署方法/算法
)

// Issuer defines a JWT issuer.
//
// JWT 签发者
type Issuer struct {
	Name        string        // 签署人名称
	Audiences   []string      // 受众列表
	IdGenerator func() string // ID 生成器

	signingKey    []byte           // 签署密钥，对于非对称加密的算法，这个字段为私钥
	signingMethod j5.SigningMethod // 签署方法/算法
	ttl           time.Duration    // 令牌有效期
}

// New creates a JWT issuer and parser using symmetric encryption algorithms,
// parameters signingMethod is the signing method/algorithm,
// parameters signingKey is the signing key, parameters ttl is the time to live of the token
//
// 创建一个使用对称加密算法的 JWT 签发者和解析器，参数 signingMethod 签署方法/算法，参数 signingKey 签署密钥，
// 参数 ttl 是令牌有效期
func New(signingMethod j5.SigningMethod, signingKey []byte, ttl time.Duration) (*Issuer, *Parser) {
	return NewIssuer(signingMethod, signingKey, ttl),
		NewParser(signingMethod, signingKey)
}

// PK creates a JWT issuer and parser using asymmetric encryption algorithms,
// parameters signingMethod is the signing method/algorithm,
// parameters privateKey is the private key, parameters publicKey is the public key, parameters ttl is the time to live of the token
//
// 创建一个使用非对称加密算法的 JWT 签发者和解析器，参数 signingMethod 签署方法/算法，参数 privateKey 私钥（仅签发者，即认证服务持有），
// 另有 publicKey 公钥分发给所有消费服务，参数 ttl 令牌有效期
func PK(signingMethod j5.SigningMethod, privateKey []byte, publicKey []byte, ttl time.Duration) (*Issuer, *Parser) {
	return NewIssuer(signingMethod, privateKey, ttl),
		NewParser(signingMethod, publicKey)
}

// NewIssuer 创建一个 JWT 签发者，参数 signingMethod 签署方法/算法，参数 signingKey 签署密钥，
// 参数 name 签发者名称，参数 audiences 消费端列表
func NewIssuer(signingMethod j5.SigningMethod, signingKey []byte, ttl time.Duration) *Issuer {
	return &Issuer{
		signingKey:    signingKey,
		signingMethod: signingMethod,
		ttl:           ttl,
		IdGenerator:   uuid.NewString,
	}
}

// DefaultIssuer creates a default JWT issuer, parameters secretKey is the key, parameters ttl is the time to live of the token
//
// 创建使用对称加密算法 HS256 的默认令牌签发者，参数 secretKey 密钥，参数 ttl 令牌有效期
func DefaultIssuer(secretKey []byte, ttl time.Duration) *Issuer {
	i := NewIssuer(DefaultSigningMethod, secretKey, ttl)
	i.Name = DefaultIssuerName
	return i
}

// DefaultPKIssuer creates a default JWT issuer, parameters privateKey is the private key, parameters ttl is the time to live of the token
//
// 创建一个使用非对称加密算法 ES256 的访问令牌签发者，参数 privateKey 私钥（仅签发者，即认证服务持有），
// 另有 publicKey 公钥分发给所有消费服务，参数 ttl 令牌有效期
func DefaultPKIssuer(privateKey []byte, ttl time.Duration) *Issuer {
	i := NewIssuer(DefaultPKSigningMethod, privateKey, ttl)
	i.Name = DefaultIssuerName
	return i
}

// Make to create a Claims object, parameters subject is the token subject,
// parameters tags are the token tags
//
// 创建一个 Claims 对象，参数 subject 令牌主题，参数 tags 令牌标签
func (i *Issuer) Make(subject string, tags ...string) (claims *Claims) {
	now := time.Now()
	expiresAt := now.Add(i.ttl)
	id := ""
	if i.IdGenerator != nil {
		id = i.IdGenerator()
	}

	claims = &Claims{
		RegisteredClaims: j5.RegisteredClaims{
			Issuer:    i.Name,
			Subject:   subject,
			ID:        id,
			Audience:  i.Audiences,
			ExpiresAt: j5.NewNumericDate(expiresAt),
			NotBefore: j5.NewNumericDate(now),
			IssuedAt:  j5.NewNumericDate(now),
		},
		Tags: tags,
	}
	return
}

// Sign to sign a JWT, returns the signed string
//
// 签署一个 JWT，返回签名后的字符串
func (i *Issuer) Sign(subject string, tags ...string) (jwt string, err error) {
	claims := i.Make(subject, tags...)
	token := j5.NewWithClaims(i.signingMethod, claims)
	return token.SignedString(i.signingKey)
}

// Generate to generate a JWT, returns a Token object
//
// 生成 JWT，返回 Token 对象
func (i *Issuer) Generate(subject string, tags ...string) (token *Token, err error) {
	v, err := i.Sign(subject, tags...)
	if err != nil {
		return nil, err
	}
	return &Token{
		Value: v,
		Ttl:   i.ttl,
	}, nil
}
