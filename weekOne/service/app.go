package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

type App struct {
	servers         []*Server
	shutdownTimeout time.Duration
	waitTimeout     time.Duration
	cbTimeout       time.Duration
	cbs             []ShutdownCallback
	//默认10秒，不知道为什么 我的windows s.shutdown 一直关闭不了服务器
	waitShotDownTimeout time.Duration
}

func WithShutdownCallback(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = append(app.cbs, cbs...)
	}
}

func WithWaitTimeout(waitTimeout time.Duration) Option {
	return func(app *App) {
		app.waitTimeout = waitTimeout
	}
}
func WithCbTimeout(cbTimeout time.Duration) Option {
	return func(app *App) {
		app.cbTimeout = cbTimeout
	}
}
func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(app *App) {
		app.shutdownTimeout = shutdownTimeout
	}
}

func NewApp(servers []*Server, opt ...Option) *App {

	a := &App{
		servers:             servers,
		shutdownTimeout:     time.Second * 30,
		waitTimeout:         time.Second * 10,
		cbTimeout:           time.Second * 3,
		cbs:                 make([]ShutdownCallback, 0, 5),
		waitShotDownTimeout: time.Second * 10,
	}
	for _, option := range opt {
		option(a)
	}
	return a
}

// StartAndServe 开启服务器 &  优雅推出
func (a *App) StartAndServe() {
	//开启服务器
	a.start()
	//监听系统信号
	a.waitSignal()
}

// 开启服务器
func (a *App) start() {
	for _, s := range a.servers {
		serv := s
		go func() {
			err := serv.Run()
			if err == http.ErrServerClosed {
				fmt.Printf("服务器%s已关闭\n", serv.name)
			} else {
				fmt.Printf("服务器%s异常退出\n", serv.name)
			}
		}()
	}
}

// 等待 系统信号
func (a *App) waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, ShutdownSignals...)
	select {
	case <-ch:
		go func() {
			select {
			case <-ch:
				fmt.Println("强制关闭服务器")
				os.Exit(1)
			case <-time.After(a.shutdownTimeout):
				fmt.Println("关闭服务器超时，强制关闭服务器")
				os.Exit(1)
			}
		}()
		a.shutdown()

	}

}

func (a *App) shutdown() {
	log.Println("开始拒绝请求")
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.waitTimeout)
	defer cancelFunc()
	for _, s := range a.servers {
		serv := s
		for _, callback := range serv.beforeCallback {
			wg.Add(1)
			callback := callback
			go func() {
				defer wg.Done()
				callback(ctx)
			}()
		}
	}
	wg.Wait()
	//在这里等待一段时间
	log.Println("开始关闭服务器")
	a.stop()
	log.Println("开始执行callback")
	a.handleCallbacks()
	log.Println("开始释放资源")
	a.close()
}

func (a *App) stop() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.waitShotDownTimeout)
	defer cancelFunc()
	done := make(chan struct{})
	go func() {
		for _, s := range a.servers {
			serv := s
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := serv.Stop(ctx)
				if err != nil {
					log.Println(err)
				}
			}()
		}
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		log.Printf("服务器关闭超时error: %v", ctx.Err())
	case <-done:
		log.Println("服务器正常关闭")
	}
}

func (a *App) handleCallbacks() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.cbTimeout)
	defer cancelFunc()
	for _, cb := range a.cbs {
		wg.Add(1)
		cb := cb
		go func() {
			defer wg.Done()
			cb(ctx)
		}()
	}
	wg.Wait()
}

func (a *App) close() {
	time.Sleep(time.Second)
	log.Println("应用关闭")
}
