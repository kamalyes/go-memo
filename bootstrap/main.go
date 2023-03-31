/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-20 09:05:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-30 10:05:22
 * @FilePath: \go-memo\bootstrap\main.go
 * @Description: go-memo 可执行服务入口，flag 对齐 redis-server 并接入 AOF 持久化
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kamalyes/go-memo/server"
)

func main() {
	port := flag.String("port", "7399", "监听端口")
	bind := flag.String("bind", "127.0.0.1", "绑定地址")
	dir := flag.String("dir", ".", "数据目录，AOF 日志存放路径")
	appendonly := flag.Bool("appendonly", true, "是否开启 AOF 持久化")
	appendfilename := flag.String("appendfilename", "appendonly.aof", "AOF 日志文件名")
	dbfilename := flag.String("dbfilename", "dump.rdb", "RDB 快照文件名，空则禁用快照")
	flag.Parse()

	addr := net.JoinHostPort(*bind, *port)
	opts := []server.Option{server.WithAddr(addr)}
	if *appendonly {
		opts = append(opts, server.WithAOF(filepath.Join(*dir, *appendfilename)))
	}
	if *dbfilename != "" {
		opts = append(opts, server.WithRDB(filepath.Join(*dir, *dbfilename)))
	}

	srv := server.New(opts...)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("memo serve: %v", err)
		}
	}()
	log.Printf("memo listening on %s (appendonly=%v rdb=%v)", addr, *appendonly, *dbfilename != "")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if err := srv.Close(); err != nil {
		log.Printf("memo shutdown: %v", err)
	}
}
