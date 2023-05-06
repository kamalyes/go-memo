/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-02 09:11:07
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-02 09:11:07
 * @FilePath: \go-memo\store\hash.go
 * @Description: 哈希类型存储操作，底层为字符串映射
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import "strconv"

// HSet 设置单个字段，复用批量方法，返回新增字段数
func (m *Memory) HSet(key, field, value string) (int, error) {
	return m.HSetMap(key, map[string]string{field: value})
}

// HSetMap 批量设置字段，返回新增字段数
func (m *Memory) HSetMap(key string, fields map[string]string) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeHash {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeHash, hash: make(map[string]string, len(fields))}
		for k, v := range fields {
			e.hash[k] = v
		}
		s.m[key] = e
		return len(fields), nil
	}
	added := 0
	for k, v := range fields {
		if _, ok := e.hash[k]; !ok {
			added++
		}
		e.hash[k] = v
	}
	return added, nil
}

// HSetNX 仅当字段不存在时写入，返回是否写入
func (m *Memory) HSetNX(key, field, value string) (bool, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeHash {
		return false, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeHash, hash: map[string]string{field: value}}
		s.m[key] = e
		return true, nil
	}
	if _, ok := e.hash[field]; ok {
		return false, nil
	}
	e.hash[field] = value
	return true, nil
}

// HGet 返回字段值，字段不存在第二个返回值为 false
func (m *Memory) HGet(key, field string) (string, bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return "", false, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return "", false, ErrWrongType
	}
	v, ok := e.hash[field]
	s.mu.RUnlock()
	return v, ok, nil
}

// HMGet 批量返回字段值，与 values 等长的 bool 标记字段是否缺失
func (m *Memory) HMGet(key string, fields ...string) ([]string, []bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return make([]string, len(fields)), make([]bool, len(fields)), nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return nil, nil, ErrWrongType
	}
	vals := make([]string, len(fields))
	oks := make([]bool, len(fields))
	for i, f := range fields {
		vals[i], oks[i] = e.hash[f]
	}
	s.mu.RUnlock()
	return vals, oks, nil
}

// HGetAll 返回全部字段值映射副本，键不存在返回空映射
func (m *Memory) HGetAll(key string) (map[string]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return map[string]string{}, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	out := make(map[string]string, len(e.hash))
	for k, v := range e.hash {
		out[k] = v
	}
	s.mu.RUnlock()
	return out, nil
}

// HDel 删除字段，返回实际删除的字段数
func (m *Memory) HDel(key string, fields ...string) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeHash {
		return 0, ErrWrongType
	}
	removed := 0
	for _, f := range fields {
		if _, ok := e.hash[f]; ok {
			delete(e.hash, f)
			removed++
		}
	}
	if len(e.hash) == 0 {
		delete(s.m, key)
	}
	return removed, nil
}

// HLen 返回字段数量，键不存在返回 0
func (m *Memory) HLen(key string) (int64, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return 0, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}
	n := int64(len(e.hash))
	s.mu.RUnlock()
	return n, nil
}

// HExists 判断字段是否存在
func (m *Memory) HExists(key, field string) (bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return false, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return false, ErrWrongType
	}
	_, ok := e.hash[field]
	s.mu.RUnlock()
	return ok, nil
}

// HKeys 返回全部字段名副本
func (m *Memory) HKeys(key string) ([]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	out := make([]string, 0, len(e.hash))
	for k := range e.hash {
		out = append(out, k)
	}
	s.mu.RUnlock()
	return out, nil
}

// HVals 返回全部字段值副本
func (m *Memory) HVals(key string) ([]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, nil
	}
	if e.typ != TypeHash {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	out := make([]string, 0, len(e.hash))
	for _, v := range e.hash {
		out = append(out, v)
	}
	s.mu.RUnlock()
	return out, nil
}

// HIncrBy 将字段值按整数递增，字段不存在视为 0，非整数值返回 ErrNotInteger
func (m *Memory) HIncrBy(key, field string, delta int64) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeHash {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeHash, hash: map[string]string{field: strconv.FormatInt(delta, 10)}}
		s.m[key] = e
		return delta, nil
	}
	cur := int64(0)
	if v, ok := e.hash[field]; ok {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, ErrNotInteger
		}
		cur = n
	}
	next := cur + delta
	e.hash[field] = strconv.FormatInt(next, 10)
	return next, nil
}