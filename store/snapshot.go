/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:36:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-17 09:36:03
 * @FilePath: \go-memo\store\snapshot.go
 * @Description: 全量快照采集，供持久化与复制使用
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// Record 单条键值快照记录
type Record struct {
	Key      string
	Value    string
	ExpireAt int64 // unix 毫秒，0 表示永不过期
}

// Snapshot 返回全量存活键值快照，过期键被过滤
func (m *Memory) Snapshot() []Record {
	now := unixMilli()
	var out []Record
	for i := range m.shards {
		s := m.shards[i]
		s.mu.RLock()
		for k, e := range s.m {
			if e.expired(now) {
				continue
			}
			out = append(out, Record{Key: k, Value: e.value, ExpireAt: e.expireAt})
		}
		s.mu.RUnlock()
	}
	return out
}
