/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-23 09:05:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 09:05:33
 * @FilePath: \go-memo\store\keys.go
 * @Description: 键枚举与统计，供 SCAN / KEYS / TYPE / DBSIZE 使用
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// Keys 返回所有存活键名，过期键被过滤
func (m *Memory) Keys() []string {
	now := unixMilli()
	var out []string
	for i := range m.shards {
		s := m.shards[i]
		s.mu.RLock()
		for k, e := range s.m {
			if e.expired(now) {
				continue
			}
			out = append(out, k)
		}
		s.mu.RUnlock()
	}
	return out
}

// Exists 判断键是否存活，过期键访问时即时剔除
func (m *Memory) Exists(key string) bool {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	s.mu.RUnlock()
	if e == nil {
		return false
	}
	if e.expired(now) {
		s.removeIfExpired(key, e, now)
		return false
	}
	return true
}

// Keyspace 返回存活键数量与带过期时间的键数量
func (m *Memory) Keyspace() (keys int64, expires int64) {
	now := unixMilli()
	for i := range m.shards {
		s := m.shards[i]
		s.mu.RLock()
		for _, e := range s.m {
			if e.expired(now) {
				continue
			}
			keys++
			if e.expireAt != 0 {
				expires++
			}
		}
		s.mu.RUnlock()
	}
	return keys, expires
}

// entryOverhead 单个键值条目除 key/value 外的估算内存开销
const entryOverhead = int64(64)

// UsedMemory 返回数据集估算内存占用字节，仅统计存活键
func (m *Memory) UsedMemory() int64 {
	now := unixMilli()
	var n int64
	for i := range m.shards {
		s := m.shards[i]
		s.mu.RLock()
		for k, e := range s.m {
			if e.expired(now) {
				continue
			}
			n += int64(len(k)) + int64(len(e.value)) + entryOverhead
		}
		s.mu.RUnlock()
	}
	return n
}
