package jwt_test

import (
	"slices"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	j1 "github.com/keepitlight/kratos/jwt"
)

func TestClaims(t *testing.T) {
	secret := "secret" // 密钥
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &j1.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			ID:        "access-token-guid", // 随机且唯一，防止重放
			Subject:   "user-open-id",
		},
		Tags: []string{"Guest"},
	})
	if s, e := token.SignedString([]byte(secret)); e != nil {
		t.Error("SignedString failed", e)
	} else if c, e := jwt.ParseWithClaims(s, &j1.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}); e != nil {
		t.Error("ParseWithClaims failed", e)
	} else if x, f := c.Claims.(*j1.Claims); !f {
		t.Error("Claims not set")
	} else if slices.Index(x.Tags, "Guest") < 0 {
		t.Error("Tags not set")
	} else if x.ID != "access-token-guid" {
		t.Error("ID not set")
	} else if x.Subject != "user-open-id" {
		t.Error("Subject not set")
	}
}
