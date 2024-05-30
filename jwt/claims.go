package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims 表示当前用户的访问凭证声明，标准字段及用途如下
//
//   - ID(jti)        string  唯一标识，一般用于凭证的安全性验证，例如防范重放攻击，或者黑名单标识
//   - Subject(sub)   string  为凭证主体（标识），例如，终端用户的标识 openid
//   - Issuer(iss)    string  为签署人（标识），谁签署了此访问凭证，此标识应在共享凭证的各方都能识别
//   - Audience(aud)  string|[]string 为授权使用的业务应用/第三方（标识），使用此凭证的应用/服务，不在此列表中的应自行拒绝
//   - IssuedAt(iat)  int     为签署时间
//   - NotBefore(nbf) int     为启用时间，不能早于签署时间，此时间之前的访问凭证将被拒绝，可选，不提供则表示立即可用
//   - ExpiresAt(exp) int     为过期时间，不能早于签署时间
type Claims struct {
	jwt.RegisteredClaims
	Tags []string `json:"tag,omitempty"` // 标签，由签署方和应用方协商实际用途
}
