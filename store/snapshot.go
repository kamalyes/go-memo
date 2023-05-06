/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:36:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-05 09:28:00
 * @FilePath: \go-memo\store\snapshot.go
 * @Description: 全量快照采集，供持久化与复制使用
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// Record 单条键值快照记录，按类型只填充对应数据字段
type Record struct {
	Key      string            // 键名
	Typ      ValueType         // 数据类型
	Str      string            // 字符串数据
	List     []string          // 列表数据
	Hash     map[string]string // 哈希数据
	Set      []string          // 集合数据
	ZSet     []ZPair           // 有序集合数据
	ExpireAt int64             // unix 毫秒，0 表示永不过期
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
			out = append(out, recordOf(k, e))
		}
		s.mu.RUnlock()
	}
	return out
}

// recordOf 将条目转换为快照记录，集合数据做深拷贝避免锁释放后被修改
func recordOf(key string, e *entry) Record {
	r := Record{Key: key, Typ: e.typ, ExpireAt: e.expireAt}
	switch e.typ {
	case TypeString:
		r.Str = e.value
	case TypeList:
		r.List = append([]string(nil), e.list...)
	case TypeHash:
		r.Hash = make(map[string]string, len(e.hash))
		for f, v := range e.hash {
			r.Hash[f] = v
		}
	case TypeSet:
		r.Set = make([]string, 0, len(e.set))
		for v := range e.set {
			r.Set = append(r.Set, v)
		}
	case TypeZSet:
		r.ZSet = make([]ZPair, 0, len(e.zset.scores))
		for member, sc := range e.zset.scores {
			r.ZSet = append(r.ZSet, ZPair{Member: member, Score: sc})
		}
	}
	return r
}
