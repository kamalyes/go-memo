/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:15:38
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 11:05:00
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

// NewRegistry 创建注册表并装载内置键值命令，连接态命令由 server 层接管
func NewRegistry() *Registry {
	r := &Registry{handlers: make(map[string]Handler), stats: NewStats()}
	r.Register("PING", handlePing)
	r.Register("GET", handleGet)
	r.Register("SET", handleSet)
	r.Register("INCRBY", handleIncrBy)
	r.Register("INCR", handleIncr)
	r.Register("DECR", handleDecr)
	r.Register("DECRBY", handleDecrBy)
	r.Register("APPEND", handleAppend)
	r.Register("STRLEN", handleStrLen)
	r.Register("MGET", handleMGet)
	r.Register("MSET", handleMSet)
	r.Register("SETEX", handleSetEX)
	r.Register("SETNX", handleSetNX)
	r.Register("DEL", handleDel)
	r.Register("EXISTS", handleExists)
	r.Register("EXPIRE", handleExpire)
	r.Register("PEXPIRE", handlePExpire)
	r.Register("TTL", handleTTL)
	r.Register("PTTL", handlePTTL)
	r.Register("PERSIST", handlePersist)
	r.Register("RENAME", handleRename)
	r.Register("DUMP", handleDump)
	r.Register("RESTORE", handleRestore)
	r.Register("TYPE", handleType)
	r.Register("DBSIZE", handleDBSize)
	r.Register("SCAN", handleScan)
	r.Register("KEYS", handleKeys)

	// 列表
	r.Register("LPUSH", handleLPush)
	r.Register("RPUSH", handleRPush)
	r.Register("LPUSHX", handleLPushX)
	r.Register("RPUSHX", handleRPushX)
	r.Register("LRANGE", handleLRange)
	r.Register("LLEN", handleLLen)
	r.Register("LPOP", handleLPop)
	r.Register("RPOP", handleRPop)
	r.Register("LREM", handleLRem)
	r.Register("LSET", handleLSet)
	r.Register("LINDEX", handleLIndex)
	r.Register("LINSERT", handleLInsert)
	r.Register("LTRIM", handleLTrim)

	// 哈希
	r.Register("HSET", handleHSet)
	r.Register("HMSET", handleHMSet)
	r.Register("HSETNX", handleHSetNX)
	r.Register("HGET", handleHGet)
	r.Register("HMGET", handleHMGet)
	r.Register("HGETALL", handleHGetAll)
	r.Register("HDEL", handleHDel)
	r.Register("HLEN", handleHLen)
	r.Register("HEXISTS", handleHExists)
	r.Register("HKEYS", handleHKeys)
	r.Register("HVALS", handleHVals)
	r.Register("HINCRBY", handleHIncrBy)

	// 集合
	r.Register("SADD", handleSAdd)
	r.Register("SREM", handleSRem)
	r.Register("SMEMBERS", handleSMembers)
	r.Register("SISMEMBER", handleSIsMember)
	r.Register("SCARD", handleSCard)
	r.Register("SPOP", handleSPop)
	r.Register("SRANDMEMBER", handleSRandMember)
	r.Register("SMOVE", handleSMove)
	r.Register("SINTER", handleSInter)
	r.Register("SUNION", handleSUnion)
	r.Register("SDIFF", handleSDiff)
	r.Register("SINTERSTORE", handleSInterStore)
	r.Register("SUNIONSTORE", handleSUnionStore)
	r.Register("SDIFFSTORE", handleSDiffStore)

	// 有序集合
	r.Register("ZADD", handleZAdd)
	r.Register("ZCARD", handleZCard)
	r.Register("ZSCORE", handleZScore)
	r.Register("ZINCRBY", handleZIncrBy)
	r.Register("ZREM", handleZRem)
	r.Register("ZRANGE", handleZRange)
	r.Register("ZREVRANGE", handleZRevRange)
	r.Register("ZRANGEBYSCORE", handleZRangeByScore)
	r.Register("ZREVRANGEBYSCORE", handleZRevRangeByScore)
	r.Register("ZRANK", handleZRank)
	r.Register("ZREVRANK", handleZRevRank)
	r.Register("ZCOUNT", handleZCount)
	r.Register("ZREMRANGEBYRANK", handleZRemRangeByRank)
	r.Register("ZREMRANGEBYSCORE", handleZRemRangeByScore)
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

// Count 返回已注册命令数量
func (r *Registry) Count() int {
	return len(r.handlers)
}

// Dispatch 按命令名分发，未知命令返回错误响应，计数由 server 层统一维护
func (r *Registry) Dispatch(st store.Store, args []string) resp.Value {
	if len(args) == 0 {
		return resp.ErrorString(msgEmptyCmd)
	}
	h, ok := r.handlers[strings.ToUpper(args[0])]
	if !ok {
		return resp.ErrorString(errUnknownCommand(args[0]))
	}
	return h(st, args)
}
