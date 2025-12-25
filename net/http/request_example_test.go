package http_test

import (
	"fmt"
	"net/http/httptest"

	http2 "github.com/keepitlight/kratos/net/http"
)

func ExampleNewRequestHelper() {
	r := httptest.NewRequest("GET", "/example", nil)
	r.Header.Set("X-Forwarded-For", "168.8.8.8, 168.8.8.9")
	h := http2.NewRequestHelper(r)
	ip, ok := h.ClientIP()
	fmt.Println(ip, ok)

	// Output:
	// 168.8.8.8 true
}
