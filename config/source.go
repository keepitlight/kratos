package config

import (
	"github.com/go-kratos/kratos/v2/config"
)

// Source return a new config source.
//
// 配置源提供者，根据已加载的本地配置 local 返回新的配置源
type Source func(local any) (config.Source, error)
