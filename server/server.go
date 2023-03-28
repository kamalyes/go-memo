/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 09:30:31
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-19 09:30:31
 * @FilePath: \go-memo\server\server.go
 * @Description: RESP 服务编排与优雅关闭
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kamalyes/go-memo/command"
	"github.com/kamalyes/go-memo/store"
)

// Server RESP 键值服务，每个连接一个 goroutine 独立读写
type Server struct {
	addr string               // 监听地址
	dbs  [DBCount]store.Store // 16 个逻辑库，各自独立键值存储
	reg  *command.Registry    // 命令注册表

	mu     sync.Mutex             // 保护连接登记表的互斥锁
	ln     net.Listener           // 已建立的监听器
	conns  map[net.Conn]*connInfo // 活动连接登记表
	nextID atomic.Int64           // 客户端自增 ID 分配器
	closed bool                   // 是否已关闭
	wg     sync.WaitGroup         // 等待在途连接退出的计数器
}

// New 创建服务，可传入装配选项
func New(opts ...Option) *Server {
	s := &Server{
		addr:  DefaultAddr,
		reg:   command.NewRegistry(),
		conns: make(map[net.Conn]*connInfo),
	}
	for i := range s.dbs {
		s.dbs[i] = store.New()
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ListenAndServe 监听并处理连接，直至 Close 被调用
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = ln.Close()
		return nil
	}
	s.ln = ln
	s.mu.Unlock()

	s.reg.Stats().SetPort(listenPort(ln))

	for {
		nc, err := ln.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			return err
		}

		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			_ = nc.Close()
			continue
		}
		ci := &connInfo{
			id:      s.nextID.Add(1),
			nc:      nc,
			addr:    nc.RemoteAddr().String(),
			created: time.Now(),
			lastCmd: time.Now(),
		}
		s.conns[nc] = ci
		s.wg.Add(1)
		s.mu.Unlock()

		s.reg.Stats().IncrTotalConn()
		s.reg.Stats().IncrConn()

		go s.serveConn(ci)
	}
}

// Close 关闭监听并断开所有连接，等待在处理中的连接退出
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	ln := s.ln
	for nc := range s.conns {
		_ = nc.Close()
	}
	s.mu.Unlock()

	if ln != nil {
		_ = ln.Close()
	}
	s.wg.Wait()
	return nil
}

// Addr 返回实际监听地址，Listener 建立后方可用
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}

// listenPort 解析监听地址的端口号，解析失败返回 0
func listenPort(ln net.Listener) int64 {
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return 0
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		return 0
	}
	return port
}
