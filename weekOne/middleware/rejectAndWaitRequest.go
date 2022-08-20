package middleware

import (
	"context"
	"geekbang-go/weekOne/service"
	"log"
	"net/http"
	"sync/atomic"
)

type Shutdown struct {
	reject   int32
	reqCount int64
	reqChan  chan struct{}
}

func NewMiddlewareShutdown() *Shutdown {
	return &Shutdown{
		reqChan: make(chan struct{}),
	}
}

func (m *Shutdown) AddReject() {
	atomic.AddInt32(&m.reject, 1)
}

func (m *Shutdown) RejectRequest() service.Handler {
	return func(ctx *service.Context) {
		//	todo: 此处代码参考 flycash/toy-web
		c1 := atomic.LoadInt32(&m.reject)
		if c1 > 0 {
			_, _ = ctx.WriteString(http.StatusServiceUnavailable, "开始拒绝请求")
			return
		}
		atomic.AddInt64(&m.reqCount, 1)
		ctx.Next()
		n := atomic.AddInt64(&m.reqCount, -1)
		if c1 > 0 && n == 0 {
			m.reqChan <- struct{}{}
		}
	}
}

func (m *Shutdown) ShutdownCallback() service.ShutdownCallback {
	return func(ctx context.Context) {
		log.Println("开始拒绝请求")
		atomic.AddInt32(&m.reject, 1)
		if atomic.LoadInt64(&m.reqCount) == 0 {
			log.Println("请求完结......")
			return
		}
		select {
		case <-ctx.Done():
			log.Println("等待请求完结超时")
			return
		case <-m.reqChan:
			log.Println("请求完结......")
			return
		}
	}
}
