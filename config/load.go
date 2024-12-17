package config

import (
	"github.com/keepitlight/kratos/runtime"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	Package = "[KeepItLight:kratos/config]" // 包名
)

var (
	_sources   []Source        // 配置源
	_providers []config.Config // 配置提供者
)

func init() {
	// !!! 注册关闭函数
	runtime.Defer(func(logger log.Logger) {
		if err := Dispose(); err != nil {
			l := log.NewHelper(logger)
			l.Error(Package, "dispose error", err)
		}
	})
}

// Dispose 调用注册的退出执行函数，由主程序调用，注意：⚠️仅执行一次
func Dispose() error {
	for _, p := range _providers {
		err := p.Close()
		if err != nil {
			return err
		}
	}
	_providers = nil
	return nil
}

// Register 注册配置提供程序，在初始化时调用
func Register(sources ...Source) {
	_sources = append(_sources, sources...)
}

// Load 加载配置
func Load[CONFIG any](prefixOfEnv, filename string) (cfg CONFIG, err error) {
	// STEP 1 加载环境变量和本地配置文件
	sources := []config.Source{
		env.NewSource(prefixOfEnv),
		file.NewSource(filename),
	}
	provider := config.New(config.WithSource(sources...))
	if err = provider.Load(); err != nil {
		return
	}

	if err = provider.Scan(&cfg); err != nil {
		return
	}
	_providers = append(_providers, provider)

	// STEP 2 加载扩展配置
	reload := false
	for _, l := range _sources {
		if s, e := l(cfg); e != nil {
			err = e
			return
		} else if s != nil {
			sources = append(sources, s)
			reload = true
		}
	}
	if !reload {
		return
	}

	if err = Dispose(); err != nil {
		return
	}

	provider = config.New(config.WithSource(sources...))
	if err = provider.Load(); err != nil {
		return
	}
	if err = provider.Scan(&cfg); err != nil {
		return
	}
	_providers = append(_providers, provider)
	return
}
