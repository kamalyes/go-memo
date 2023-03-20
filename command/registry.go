/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:15:38
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 09:15:38
 * @FilePath: \go-memo\command\registry.go
 * @Description: 命令注册表与分发器
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strings"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// Handler 命令处理器，接收存储后端与完整参数列表返回响应值
type Handler func(st store.Store, args []string) resp.Value

// Registry 命令注册表，命令名大小写不敏感
type Registry struct {
	handlers map[string]Handler // 命令名到处理器的映射
	stats    *Stats             // 运行期统计状态
}

// NewRegistry 创建注册表并装载内置命令
func NewRegistry() *Registry {
	r := &Registry{handlers: make(map[string]Handler), stats: NewStats()}
	r.Register("PING", handlePing)
	r.Register("SELECT", handleSelect)
	r.Register("INFO", r.handleInfo)
	r.Register("GET", handleGet)
	r.Register("SET", handleSet)
	r.Register("INCRBY", handleIncrBy)
	r.Register("DEL", handleDel)
	r.Register("EXPIRE", handleExpire)
	r.Register("PEXPIRE", handlePExpire)
	r.Register("TTL", handleTTL)
	r.Register("PTTL", handlePTTL)
	r.Register("TYPE", handleType)
	r.Register("DBSIZE", handleDBSize)
	r.Register("SCAN", handleScan)
	r.Register("KEYS", handleKeys)
	return r
}

// Register 注册命令处理器
func (r *Registry) Register(name string, h Handler) {
	r.handlers[strings.ToUpper(name)] = h
}

// Stats 返回注册表关联的运行期统计状态
func (r *Registry) Stats() *Stats {
	return r.stats
}

// Dispatch 按命令名分发，未知命令返回错误响应
func (r *Registry) Dispatch(st store.Store, args []string) resp.Value {
	if len(args) == 0 {
		return resp.ErrorString(msgEmptyCmd)
	}
	r.stats.IncrCmd()
	h, ok := r.handlers[strings.ToUpper(args[0])]
	if !ok {
		return resp.ErrorString(errUnknownCommand(args[0]))
	}
	return h(st, args)
}
