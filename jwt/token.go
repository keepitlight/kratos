package jwt

import "time"

const (
	Bearer        = "Bearer"        // 令牌类型，通常是 Bearer
	Authorization = "Authorization" // HTTP 头中的授权字段
)

// Token 返回的 Token 信息
type Token struct {
	Value string        `json:"value"` // 令牌值
	Ttl   time.Duration `json:"ttl"`   // 持续时间，单位秒
}
