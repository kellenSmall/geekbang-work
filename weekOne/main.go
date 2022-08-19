package main

import (
	"context"
	"fmt"
	"geekbang-go/weekOne/service"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	s1 := service.NewServer("business", "localhost:8080")
	s1.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello business"))
	}))

	s2 := service.NewServer("admin", "localhost:8081")
	s2.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello admin"))
	}))

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
