/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-30 09:12:35
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-30 09:12:35
 * @FilePath: \go-memo\server\persist.go
 * @Description: AOF 持久化接线，写命令追加与启动回放
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"strings"

	"github.com/kamalyes/go-memo/persistence"
)

// writeCommands 需要追加到 AOF 的状态变更命令集合，SELECT 用于回放时路由逻辑库
var writeCommands = map[string]bool{
	"SET":      true,
	"SETNX":    true,
	"SETEX":    true,
	"MSET":     true,
	"INCR":     true,
	"DECR":     true,
	"INCRBY":   true,
	"DECRBY":   true,
	"APPEND":   true,
	"DEL":      true,
	"EXPIRE":   true,
	"PEXPIRE":  true,
	"PERSIST":  true,
	"RENAME":   true,
	"RESTORE":  true,
	"SELECT":   true,
	"FLUSHDB":  true,
	"FLUSHALL": true,
}

// isWriteCommand 判断命令是否为需要久化的状态变更命令
func isWriteCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return writeCommands[strings.ToUpper(args[0])]
}

// initPersistence 回放已有 AOF 日志并打开追加句柄，未启用时直接返回
func (s *Server) initPersistence() error {
	if s.aofPath == "" {
		return nil
	}
	ci := &connInfo{}
	if err := persistence.Replay(s.aofPath, func(args []string) error {
		s.dispatch(ci, args)
		return nil
	}); err != nil {
		return err
	}
	aof, err := persistence.OpenAOF(s.aofPath)
	if err != nil {
		return err
	}
	s.aof = aof
	return nil
}

// appendAOF 将状态变更命令追加到持久化日志
func (s *Server) appendAOF(args []string) error {
	if s.aof == nil || !isWriteCommand(args) {
		return nil
	}
	return s.aof.Append(args)
}
