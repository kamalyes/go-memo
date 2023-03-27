/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-20 09:05:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-20 09:05:33
 * @FilePath: \go-memo\bootstrap\main.go
 * @Description: go-memo 可执行服务入口，装配服务并优雅关闭
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kamalyes/go-memo/server"
)

func main() {
	addr := flag.String("addr", server.DefaultAddr, "监听地址")
	flag.Parse()

	srv := server.New(server.WithAddr(*addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("memo serve: %v", err)
		}
	}()
	log.Printf("memo listening on %s", *addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if err := srv.Close(); err != nil {
		log.Printf("memo shutdown: %v", err)
	}
}
