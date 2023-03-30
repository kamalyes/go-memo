/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 09:18:22
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-19 09:18:22
 * @FilePath: \go-memo\server\options.go
 * @Description: 服务装配选项定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import "github.com/kamalyes/go-memo/store"

// Option 服务装配选项
type Option func(*Server)

// WithAddr 设置监听地址
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithStore 注入存储后端作为 0 号逻辑库，默认使用内存实现
func WithStore(st store.Store) Option {
	return func(s *Server) {
		s.dbs[0] = st
	}
}

// WithAOF 开启 AOF 持久化，写命令追加到指定文件并在启动时回放
func WithAOF(path string) Option {
	return func(s *Server) {
		s.aofPath = path
	}
}
