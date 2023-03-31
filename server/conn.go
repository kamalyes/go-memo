/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 09:35:13
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 11:20:00
 * @FilePath: \go-memo\server\conn.go
 * @Description: 连接读写循环与命令分发
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/kamalyes/go-memo/resp"
)

// connInfo 单条活动连接的元信息，含客户端身份与所选逻辑库
type connInfo struct {
	id      int64     // 客户端自增 ID
	nc      net.Conn  // 底层连接
	addr    string    // 对端地址
	name    string    // 客户端名称，由 CLIENT SETNAME 设置
	db      int       // 当前所选逻辑库索引
	created time.Time // 建立连接时刻
	lastCmd time.Time // 最近一次命令时刻
	quit    bool      // 是否收到 QUIT 标记
}

// serveConn 处理单条连接，循环读取命令、分发并写回响应
func (s *Server) serveConn(ci *connInfo) {
	defer s.wg.Done()
	defer s.dropConn(ci.nc)
	defer ci.nc.Close()
	defer s.reg.Stats().DecrConn()

	br := bufio.NewReader(ci.nc)
	w := resp.NewWriter(ci.nc)
	for {
		args, err := resp.ReadCommand(br)
		if err != nil {
			return
		}
		ci.lastCmd = time.Now()
		reply := s.dispatch(ci, args)
		if reply.Type != resp.TypeError {
			if err := s.appendAOF(args); err != nil {
				return
			}
		}
		if err := w.WriteValue(reply); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}
		if ci.quit {
			return
		}
	}
}

// dispatch 分发单条命令，优先处理连接态与服务器级命令，其余落到当前逻辑库
func (s *Server) dispatch(ci *connInfo, args []string) resp.Value {
	s.reg.Stats().IncrCmd()
	if len(args) == 0 {
		return resp.ErrorString(msgEmptyCmd)
	}
	switch strings.ToUpper(args[0]) {
	case "SELECT":
		return s.cmdSelect(ci, args)
	case "QUIT":
		return s.cmdQuit(ci, args)
	case "CLIENT":
		return s.cmdClient(ci, args)
	case "FLUSHDB":
		return s.cmdFlushDB(ci, args)
	case "FLUSHALL":
		return s.cmdFlushAll(args)
	case "INFO":
		return s.cmdInfo(args)
	case "CONFIG":
		return s.cmdConfig(args)
	case "COMMAND":
		return s.cmdCommand(args)
	case "HELLO":
		return s.cmdHello(ci, args)
	case "SAVE":
		return s.cmdSave(args)
	case "BGSAVE":
		return s.cmdBGSave(args)
	case "CLUSTER":
		return s.cmdCluster(args)
	}
	return s.reg.Dispatch(s.dbs[ci.db], args)
}

// dropConn 从连接登记表移除已关闭的连接
func (s *Server) dropConn(nc net.Conn) {
	s.mu.Lock()
	delete(s.conns, nc)
	s.mu.Unlock()
}
