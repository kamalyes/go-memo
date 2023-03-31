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

// initPersistence 先回放 RDB 快照再回放 AOF 日志，最后打开追加句柄
func (s *Server) initPersistence() error {
	if s.rdbPath != "" {
		if err := s.replayFile(s.rdbPath); err != nil {
			return err
		}
	}
	if s.aofPath == "" {
		return nil
	}
	if err := s.replayFile(s.aofPath); err != nil {
		return err
	}
	aof, err := persistence.OpenAOF(s.aofPath)
	if err != nil {
		return err
	}
	s.aof = aof
	return nil
}

// replayFile 回放命令流文件，SELECT 与写命令经 dispatch 路由到对应逻辑库
func (s *Server) replayFile(path string) error {
	ci := &connInfo{}
	return persistence.Replay(path, func(args []string) error {
		s.dispatch(ci, args)
		return nil
	})
}

// appendAOF 将状态变更命令追加到持久化日志
func (s *Server) appendAOF(args []string) error {
	if s.aof == nil || !isWriteCommand(args) {
		return nil
	}
	return s.aof.Append(args)
}
