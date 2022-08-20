package main

import (
	"context"
	"fmt"
	"geekbang-go/weekOne/middleware"
	"geekbang-go/weekOne/service"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	shutdown := middleware.NewMiddlewareShutdown()
	s1 := service.NewServer("business", "localhost:8080", service.WithBeforeCallback(shutdown.ShutdownCallback()))
	s1.Use(shutdown.RejectRequest())
	s1.Get("/", func(ctx *service.Context) {
		_, _ = ctx.WriteString(http.StatusOK, "hello business")
	})

	shutdown2 := middleware.NewMiddlewareShutdown()
	s2 := service.NewServer("admin", "localhost:8081", service.WithBeforeCallback(shutdown2.ShutdownCallback()))
	s2.Use(shutdown2.RejectRequest())
	s2.Get("/", func(ctx *service.Context) {
		_, _ = ctx.WriteString(http.StatusOK, "hello admin")
	})
	app := service.NewApp([]*service.Server{s1, s2}, service.WithShutdownCallback(StoreCacheToDBCallback))
	app.StartAndServe()
}

func StoreCacheToDBCallback(ctx context.Context) {
	done := make(chan struct{}, 1)
	go func() {
		log.Println("刷新缓存中......")
		rand.Seed(time.Now().Unix())
		r := rand.Int63n(4)
		fmt.Println(r)
		time.Sleep(time.Second * time.Duration(r))
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		log.Println("刷新缓存超时。。。")
	case <-done:
		log.Println("缓存刷新到Db")
	}
}
