/*
 Q: 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。

 A: 解题思路，

 1. 知识点
    - waitgroup: sync.WaitGroup
	- goroutine: 需要知道"如何结束"和"怎么结束"
    - errgroup

	waitgroup 基本使用：
	var wg sync.WaitGroup 声明一个 waitgroup 变量（相当于计数器）
	开启一个 goroutine 时，先使用 wg.Add(1) 增加一个计数
	完成时，使用 defer wg.Done() 减少一个计数
	wg.Wait() 直到计数为 0 时，结束线程

*/

package main

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"os/signal"

	"github.com/pkg/errors"

	"golang.org/x/sync/errgroup"
)

func homework() {
	// 创建带 context 的 errgroup
	g, ctx := errgroup.WithContext(context.Background())
	shutdown := make(chan int, 0)
	quit := make(chan os.Signal, 0)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 路由创建
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("<h1>Hello, XavierChan</h1>"))
	})
	mux.HandleFunc("/shutdown", func(rw http.ResponseWriter, r *http.Request) {
		shutdown <- 1
	})

	server := &http.Server{
		Addr:           ":6001",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// g1 Http Server
	g.Go(func() error {
		return server.ListenAndServe()
	})

	// g2
	g.Go(func() error {
		select {
		case <-ctx.Done():
			log.Println("errgroup exit...")
		case <-shutdown:
			log.Println("server will out...")
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		log.Println("shutting down server...")
		return server.Shutdown(timeoutCtx)
	})

	// g3
	g.Go(func() error {

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-quit:
			return errors.Errorf("get os signal: %v", sig)
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Println("exit...")
	}
}

func main() {
	homework()
}
