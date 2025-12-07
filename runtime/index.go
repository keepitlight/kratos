package runtime

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
)

const (
	Package = "kratos/runtime"
)

var (
	runtime = &Runtime{once: sync.Once{}, exit: sync.Once{}}
)

// Runtime 运行时
type Runtime struct {
	readies  []func() error // 准备程序
	defers   []func()       // 延迟程序
	routines []Routine      // 伴生协程

	appInfo   kratos.AppInfo     // 当前的主程序信息，仅在主程序运行后被设置为有效信息
	registrar registry.Registrar // 当前的注册中心
	build     string             // 构建时间信息
	commit    string             // 提交版本信息
	uptime    time.Time          // 程序开始的运行时间

	once sync.Once
	exit sync.Once
}

// start 启动运行时，⚠️仅运行一次
func (r *Runtime) start(
	ctx context.Context,
	appInfo kratos.AppInfo,
	registrar registry.Registrar,
	build, commit string,
	uptime time.Time) (msg <-chan error, err error, ok bool) {

	r.once.Do(func() {
		// msg = make(chan error)
		r.appInfo = appInfo
		r.registrar = registrar
		r.build = build
		r.commit = commit
		r.uptime = uptime
		// 预加载的函数
		for _, ready := range r.readies {
			err = ready()
			if err != nil {
				return
			}
		}
		msg = r.run(ctx)
		ok = true
	})
	return
}

func (r *Runtime) State() (build, commit string, uptime time.Time) {
	return r.build, r.commit, r.uptime
}

func (r *Runtime) AppInfo() kratos.AppInfo {
	return r.appInfo
}

func (r *Runtime) Registrar() registry.Registrar {
	return r.registrar
}

// Preload 指定在主程序启动时执行的函数，此方法要在 init 函数中调用，否则可能会被忽略
func (r *Runtime) preload(f func() error) {
	r.readies = append(r.readies, f)
}

// Defer 指定在主程序退出时执行的函数，全局延迟函数
func (r *Runtime) deferIt(f func()) {
	r.defers = append(r.defers, f)
}

// Co 增加伴生协程，以在主协程启动时执行，伴生协程退出或异常不影响主协程，
// 但主协程退出或异常，伴生协程收到上下文退出通知要主动退出，注意，在 init 中调用，
// 否则会被忽略
func (r *Runtime) co(routines ...Routine) {
	routines = slices.DeleteFunc(routines, func(r Routine) bool { return r == nil })
	r.routines = append(r.routines, routines...)
}

// run 执行所有注册的伴生协程，与主协程协同运行，伴生协程退出或异常不影响主协程，
// 返回带缓冲的消息通道，返回通道关闭表示所有伴生协程正常退出。
// 但主协程退出或异常，伴生协程收到通知要主动退出
func (r *Runtime) run(ctx context.Context) <-chan error {
	var c = make(chan error, len(r.routines))
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(r.routines))
		for _, ro := range r.routines {
			go func(routine Routine) {
				defer func() {
					wg.Done()

					if p := recover(); p != nil {
						// 意外的 panic，打印堆栈信息
						c <- fmt.Errorf("[kratos/runtime]panic catch, routine throw error:\n%v\n\n", p)
					}
				}()

				if e := routine(ctx); e != nil {
					c <- e
				}
			}(ro)
		}
		wg.Wait()
		close(c)
	}()
	return c
}

func (r *Runtime) dispose() {
	r.exit.Do(func() {
		for _, def := range r.defers {
			def()
		}
	})
}

// Co 增加伴生协程，以在主协程启动时执行，伴生协程退出或异常不影响主协程，
// 但主协程退出或异常，伴生协程收到通知要主动退出，注意，在 init 中调用，否则会被忽略
func Co(routines ...Routine) {
	runtime.co(routines...)
}

// Start 启动运行时，⚠️仅运行一次
// 如需关闭伴生协程，传递可以取消的上下文，通过上下文关闭伴生协程
func Start(
	ctx context.Context,
	appInfo kratos.AppInfo,
	registrar registry.Registrar,
	build, commit string,
	uptime time.Time) (channel <-chan error, err error, ok bool) {
	return runtime.start(ctx, appInfo, registrar, build, commit, uptime)
}

func State() (
	appInfo kratos.AppInfo,
	registrar registry.Registrar,
	build, commit string,
	uptime time.Time,
) {
	return runtime.appInfo, runtime.registrar, runtime.build, runtime.commit, runtime.uptime
}

// Preload 指定在主程序启动时执行的函数，此方法要在 init 函数中调用，否则会被忽略
func Preload(f func() error) {
	runtime.preload(f)
}

// Defer 指定在主程序退出时执行的函数
func Defer(f func()) {
	runtime.deferIt(f)
}

func Current() (current *Runtime) {
	return runtime
}

func Dispose() {
	runtime.dispose()
}
