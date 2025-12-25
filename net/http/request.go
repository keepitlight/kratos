package http

import (
	"context"
	"net"
	"strings"

	gnet "github.com/keepitlight/golang/net"

	"github.com/go-kratos/kratos/v2/transport/http"
)

var (
	ClientAndProxyIPHeaders = []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Forwarded",
		"HTTP_CLIENT_IP",
		"HTTP_X_FORWARDED_FOR",
		"Proxy-Client-IP",
		"WL-Proxy-Client-IP",
	}
)

// RequestHelper represents a helper to get http server header
//
// 获取 http server 的 header 帮助器
type RequestHelper struct {
	request *http.Request
}

func NewRequestHelper(req *http.Request) *RequestHelper {
	if req == nil {
		return nil
	}
	h := &RequestHelper{
		request: req,
	}
	return h
}

func FromContext(ctx context.Context) *RequestHelper {
	req, ok := http.RequestFromServerContext(ctx)
	if !ok || req == nil {
		return nil
	}
	return NewRequestHelper(req)
}

func (h *RequestHelper) IPs() (ip []net.IP) {
	if h.request == nil || h.request.Header == nil {
		return
	}
	for _, header := range ClientAndProxyIPHeaders {
		if vs := h.request.Header.Values(header); len(vs) > 0 {
			for _, v := range vs {
				if v == "" {
					continue
				}
				v = strings.ReplaceAll(v, " ", "")
				v = strings.ReplaceAll(v, ";", ",")
				if strings.IndexRune(v, ',') >= 0 {
					vv := strings.Split(v, ",")
					v = vv[0] // get first
				}
				if x := net.ParseIP(v); gnet.IsPublic(x) {
					ip = append(ip, x)
				}
			}
		}
	}
	return
}

func (h *RequestHelper) ClientIP() (ip net.IP, found bool) {
	if h.request == nil || h.request.Header == nil {
		return
	}
	for _, header := range ClientAndProxyIPHeaders {
		if vs := h.request.Header.Values(header); len(vs) > 0 {
			for _, v := range vs {
				if v == "" {
					continue
				}
				v = strings.ReplaceAll(v, " ", "")
				v = strings.ReplaceAll(v, ";", ",")
				if strings.IndexRune(v, ',') >= 0 {
					vv := strings.Split(v, ",")
					v = vv[0] // get first
				}
				if x := net.ParseIP(v); gnet.IsPublic(x) {
					return x, true
				}
			}
		}
	}
	return
}

// Header returns a list of values for any header name in the given order
//
// 根据提供的任一头名称，返回该头对应的值列表。如果头不存在，则返回 nil 和 false。
func (h *RequestHelper) Header(headers ...string) ([]string, bool) {
	if len(headers) < 1 || h.request == nil || h.request.Header == nil {
		return nil, false
	}
	for _, header := range headers {
		if v := h.request.Header.Values(header); len(v) > 0 {
			return v, true
		}
	}
	return nil, false
}
