/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:23:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-17 09:23:55
 * @FilePath: \go-memo\store\memory.go
 * @Description: 分片内存键值存储，过期键访问时懒删除
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// ErrNotInteger 键值无法解析为整数
var ErrNotInteger = errors.New("value is not an integer or out of range")

// entry 单个键的存储单元
type entry struct {
	value    string
	expireAt int64 // unix 毫秒，0 表示永不过期
}

// expired 判断键是否在指定时刻已过期
func (e *entry) expired(now int64) bool {
	return e.expireAt != 0 && e.expireAt <= now
}

// Memory 分片内存键值存储，哈希分散到 256 个独立锁降低竞争
type Memory struct {
	shards [shardCount]*shard // 分片数组，哈希分散到独立锁
}

// shard 单个分片，读写锁允许并发读、串行写
type shard struct {
	mu sync.RWMutex      // 分片读写锁
	m  map[string]*entry // 分片内键值映射
}

// New 创建内存键值存储
func New() *Memory {
	mem := &Memory{}
	for i := range mem.shards {
		mem.shards[i] = &shard{m: make(map[string]*entry)}
	}
	return mem
}

// Get 读取键值，键不存在或已过期返回 false
func (m *Memory) Get(key string) (string, bool) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	s.mu.RUnlock()
	if e == nil {
		return "", false
	}
	if e.expired(now) {
		s.removeIfExpired(key, e, now)
		return "", false
	}
	return e.value, true
}

// Set 写入键值，覆盖过期时间
func (m *Memory) Set(key string, value string) {
	s := m.shardOf(key)
	s.mu.Lock()
	s.m[key] = &entry{value: value}
	s.mu.Unlock()
}

// IncrBy 将键值按整数递增
func (m *Memory) IncrBy(key string, delta int64) (int64, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.Lock()
	defer s.mu.Unlock()

	cur := int64(0)
	if e, ok := s.m[key]; ok {
		if e.expired(now) {
			delete(s.m, key)
		} else {
			n, err := strconv.ParseInt(e.value, 10, 64)
			if err != nil {
				return 0, ErrNotInteger
			}
			cur = n
		}
	}
	next := cur + delta
	s.m[key] = &entry{value: strconv.FormatInt(next, 10)}
	return next, nil
}

// Del 删除键，返回删除的存活键数量
func (m *Memory) Del(keys ...string) int {
	now := unixMilli()
	n := 0
	for _, key := range keys {
		s := m.shardOf(key)
		s.mu.Lock()
		if e, ok := s.m[key]; ok {
			delete(s.m, key)
			if !e.expired(now) {
				n++
			}
		}
		s.mu.Unlock()
	}
	return n
}

// Expire 为存活键设置过期时间
func (m *Memory) Expire(key string, expireAt int64) bool {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[key]
	if !ok {
		return false
	}
	if e.expired(now) {
		delete(s.m, key)
		return false
	}
	e.expireAt = expireAt
	return true
}

// TTL 返回剩余存活毫秒数
func (m *Memory) TTL(key string) int64 {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	s.mu.RUnlock()
	if e == nil {
		return TTLNotExist
	}
	if e.expireAt == 0 {
		return TTLNoExpire
	}
	remaining := e.expireAt - now
	if remaining <= 0 {
		s.removeIfExpired(key, e, now)
		return TTLNotExist
	}
	return remaining
}

// removeIfExpired 仅在键仍为同一对象且未过期时删除，避免覆盖并发写入
func (s *shard) removeIfExpired(key string, e *entry, now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.m[key]; ok && cur == e && cur.expired(now) {
		delete(s.m, key)
	}
}

func (m *Memory) shardOf(key string) *shard {
	return m.shards[fnv32(key)%shardCount]
}

func fnv32(key string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

func unixMilli() int64 {
	return time.Now().UnixMilli()
}
