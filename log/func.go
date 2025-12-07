package log

import (
	"log/slog"
	"strings"
)

const (
	DefaultRedactionValue = "<*REDACTED*>"
	Wildcard              = "*"
)

// Matcher 用于匹配指定的日志键。
type Matcher interface {
	Match(key string, groups ...string) bool
}

type MatcherFunc func(key string, groups ...string) bool

func (m MatcherFunc) Match(key string, groups ...string) bool {
	return m(key, groups...)
}

// KeyOf 用于匹配指定的日志键，忽略组，区分大小写。
func KeyOf(keys ...string) Matcher {
	var items map[string]struct{} = nil
	for _, k := range keys {
		items[k] = struct{}{}
	}
	return MatcherFunc(func(key string, groups ...string) bool {
		if len(items) > 0 {
			// 1. 匹配处理
			if _, found := items[key]; found {
				return true
			}
		}
		return false
	})
}

// FoldKeyOf 用于匹配指定的日志键，忽略组，不区分大小写。
func FoldKeyOf(keys ...string) Matcher {
	var items map[string]struct{} = nil
	for _, k := range keys {
		items[strings.ToLower(k)] = struct{}{}
	}
	return MatcherFunc(func(key string, groups ...string) bool {
		if len(items) > 0 {
			// 1. 匹配处理
			if _, found := items[strings.ToLower(key)]; found {
				return true
			}
		}
		return false
	})
}

// GroupOf 用于匹配指定组下的日志键，区分大小写。
// groups 参数指示指定组的键的路径，keys 参数指定组下的日志键的集合。
// groups 支持通配符，仅限最后一段组键，使用 * 表示通配，如：`["user_ip", "*"]`。
// 通配匹配模式下，匹配的日志键必须以指定的前缀开头。
func GroupOf(groups []string, keys ...string) Matcher {
	var items map[string]struct{} = nil
	for _, k := range keys {
		items[k] = struct{}{}
	}
	var wildcard = false
	l := len(groups)
	if l > 0 {
		wildcard = groups[l-1] == Wildcard
		if wildcard {
			groups = groups[:l-1]
			l--
		}
	}
	return MatcherFunc(func(key string, targets ...string) bool {
		t := len(targets)
		if l > 0 {
			// 1. 匹配组
			if !wildcard && t != l {
				return false
			}
			if t < l {
				return false
			}
			for i, g := range groups {
				s := targets[i]
				if g != s {
					return false
				}
			}
		} else if t > 0 {
			return false
		}
		// 2. 匹配日志键
		if len(items) > 0 {
			if _, found := items[key]; found {
				return true
			}
		}
		return false
	})
}

// FoldGroupOf 用于匹配指定组下的日志键，不区分大小写。
// groups 参数指示指定组的键的路径，keys 参数指定组下的日志键的集合。
// groups 支持通配符，仅限最后一段组键，使用 * 表示通配，如：`["user_ip", "*"]`。
// 通配匹配模式下，匹配的日志键必须以指定的前缀开头。
func FoldGroupOf(groups []string, keys ...string) Matcher {
	var items map[string]struct{} = nil
	for _, k := range keys {
		items[strings.ToLower(k)] = struct{}{}
	}
	var wildcard = false
	l := len(groups)
	if l > 0 {
		wildcard = groups[l-1] == Wildcard
		if wildcard {
			groups = groups[:l-1]
			l--
		}
	}
	var gs []string
	for _, g := range groups {
		gs = append(gs, strings.ToLower(g))
	}
	return MatcherFunc(func(key string, targets ...string) bool {
		t := len(targets)
		if l > 0 {
			// 1. 匹配组
			if !wildcard && t != l {
				return false
			}
			if t < l {
				return false
			}
			for i, g := range gs {
				s := targets[i]
				if !strings.EqualFold(g, s) {
					return false
				}
			}
		} else if t > 0 {
			return false
		}
		// 2. 匹配日志键
		if len(items) > 0 {
			if _, found := items[strings.ToLower(key)]; found {
				return true
			}
		}
		return false
	})
}

// Rewrite 创建一个对日志数据进行重写处理的函数。
// rewrite 参数用于对日志数据进行重写，matchers 指示要匹配的日志项目的集合。
func Rewrite(rewrite func(value slog.Value) slog.Value, matchers ...Matcher) func(groups []string, a slog.Attr) (slog.Attr, bool) {
	return func(groups []string, a slog.Attr) (slog.Attr, bool) {
		if len(matchers) > 0 && a.Value.Kind() == slog.KindString {
			// 确保只对字符串值进行脱敏，避免破坏数字或布尔值
			// 这里我们直接用配置的脱敏值替换整个 Value
			// 2. 脱敏处理
			for _, matcher := range matchers {
				if matcher.Match(a.Key, groups...) {
					return slog.Attr{
						Key:   a.Key,
						Value: rewrite(a.Value),
					}, true
				}
			}
		}
		return a, false
	}
}

// Redact 创建一个对字符串日志数据进行脱敏处理的函数。
func Redact(redaction string, matchers ...Matcher) func(groups []string, a slog.Attr) (slog.Attr, bool) {
	r := DefaultRedactionValue
	if redaction != "" {
		r = redaction
	}
	return func(groups []string, a slog.Attr) (slog.Attr, bool) {
		if len(matchers) > 0 && a.Value.Kind() == slog.KindString {
			// 确保只对字符串值进行脱敏，避免破坏数字或布尔值
			// 这里我们直接用配置的脱敏值替换整个 Value
			// 2. 脱敏处理
			for _, matcher := range matchers {
				if matcher.Match(a.Key, groups...) {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.StringValue(r),
					}, true
				}
			}
		}
		return a, false
	}
}

func DefaultRedact() func(groups []string, a slog.Attr) (slog.Attr, bool) {
	return Redact(
		DefaultRedactionValue,
		KeyOf("password", "secret", "token", "access_token", "refresh_token"),
	)
}

// Rename 创建一个日志过滤器，用于将指定的日志键重命名。
// keys 存储需要过滤的日志键集合，keys 参数的键为日志键匹配器，keys 参数的值为新的日志键名。
func Rename(keys map[Matcher]string) func(groups []string, a slog.Attr) (slog.Attr, bool) {
	return func(groups []string, a slog.Attr) (slog.Attr, bool) {
		if len(keys) > 0 {
			for matcher, newKey := range keys {
				if matcher.Match(a.Key, groups...) {
					return slog.Attr{
						Key:   newKey,
						Value: a.Value,
					}, true
				}
			}
		}
		return a, false
	}
}

// Ignore 创建一个日志过滤器，用于忽略指定的日志键。
// matchers 存储需要过滤的日志键匹配器的集合。
func Ignore(matchers ...Matcher) func(groups []string, a slog.Attr) (slog.Attr, bool) {
	return func(groups []string, a slog.Attr) (slog.Attr, bool) {
		if len(matchers) > 0 {
			for _, matcher := range matchers {
				if matcher.Match(a.Key, groups...) {
					return slog.Attr{}, true
				}
			}
		}
		return a, false
	}
}

// Compose 创建一个日志过滤器，用于过滤掉指定的日志键。
// addons 是一个日志过滤器函数列表。
func Compose(addons ...func(groups []string, a slog.Attr) (r slog.Attr, ok bool)) func([]string, slog.Attr) slog.Attr {
	// 返回实际的回调函数
	return func(groups []string, a slog.Attr) slog.Attr {
		var v = a
		for _, addon := range addons {
			if addon == nil {
				continue
			}
			if r, ok := addon(groups, v); ok {
				return r
			} else {
				v = r
			}
		}
		return a
	}
}
