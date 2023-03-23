/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 09:35:13
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-19 09:35:13
 * @FilePath: \go-memo\server\conn.go
 * @Description: 连接读写循环处理
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"bufio"
	"net"

	"github.com/kamalyes/go-memo/resp"
)

// serveConn 处理单条连接，循环读取命令、分发并写回响应
func (s *Server) serveConn(nc net.Conn) {
	defer s.wg.Done()
	defer s.dropConn(nc)
	defer nc.Close()
	defer s.reg.Stats().DecrConn()

	br := bufio.NewReader(nc)
	w := resp.NewWriter(nc)
	for {
		args, err := resp.ReadCommand(br)
		if err != nil {
			return
		}
		reply := s.reg.Dispatch(s.st, args)
		if err := w.WriteValue(reply); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}
	}
}

// dropConn 从连接登记表移除已关闭的连接
func (s *Server) dropConn(nc net.Conn) {
	s.mu.Lock()
	delete(s.conns, nc)
	s.mu.Unlock()
}
