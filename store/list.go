/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-01 09:02:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-01 09:02:11
 * @FilePath: \go-memo\store\list.go
 * @Description: 列表类型存储操作，底层为切片
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// LPush 将多个值依次插入列表头部，返回插入后的列表长度
func (m *Memory) LPush(key string, values ...string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeList {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeList}
		s.m[key] = e
	}
	dst := make([]string, 0, len(values)+len(e.list))
	for i := len(values) - 1; i >= 0; i-- {
		dst = append(dst, values[i])
	}
	e.list = append(dst, e.list...)
	return int64(len(e.list)), nil
}

// RPush 将多个值依次插入列表尾部，返回插入后的列表长度
func (m *Memory) RPush(key string, values ...string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeList {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeList, list: make([]string, 0, len(values))}
		s.m[key] = e
	}
	e.list = append(e.list, values...)
	return int64(len(e.list)), nil
}

// LPushX 仅当键存在时插入头部，返回插入后的列表长度，键不存在返回 0
func (m *Memory) LPushX(key string, values ...string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeList {
		return 0, ErrWrongType
	}
	dst := make([]string, 0, len(values)+len(e.list))
	for i := len(values) - 1; i >= 0; i-- {
		dst = append(dst, values[i])
	}
	e.list = append(dst, e.list...)
	return int64(len(e.list)), nil
}

// RPushX 仅当键存在时插入尾部，返回插入后的列表长度，键不存在返回 0
func (m *Memory) RPushX(key string, values ...string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeList {
		return 0, ErrWrongType
	}
	e.list = append(e.list, values...)
	return int64(len(e.list)), nil
}

// LRange 返回列表指定区间的元素，区间越界或空洞返回空切片
func (m *Memory) LRange(key string, start, stop int) ([]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, nil
	}
	if e.typ != TypeList {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	from, to, ok := rangeIndices(start, stop, len(e.list))
	if !ok {
		s.mu.RUnlock()
		return nil, nil
	}
	out := make([]string, to-from)
	copy(out, e.list[from:to])
	s.mu.RUnlock()
	return out, nil
}

// LLen 返回列表长度，键不存在返回 0
func (m *Memory) LLen(key string) (int64, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return 0, nil
	}
	if e.typ != TypeList {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}
	n := int64(len(e.list))
	s.mu.RUnlock()
	return n, nil
}

// LPop 从列表头部弹出 count 个元素，existed 标识键是否存在且为列表
func (m *Memory) LPop(key string, count int) ([]string, bool, error) {
	return m.pop(key, count, true)
}

// RPop 从列表尾部弹出 count 个元素，existed 标识键是否存在且为列表
func (m *Memory) RPop(key string, count int) ([]string, bool, error) {
	return m.pop(key, count, false)
}

// pop 统一实现头部或尾部弹出，弹出后列表为空则删除键
func (m *Memory) pop(key string, count int, left bool) ([]string, bool, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, false, nil
	}
	if e.typ != TypeList {
		return nil, false, ErrWrongType
	}
	if count <= 0 {
		count = 1
	}
	if n := len(e.list); n > 0 {
		if count > n {
			count = n
		}
	}
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if left {
			out = append(out, e.list[0])
			e.list = e.list[1:]
		} else {
			last := len(e.list) - 1
			out = append(out, e.list[last])
			e.list = e.list[:last]
		}
		if len(e.list) == 0 {
			break
		}
	}
	if len(e.list) == 0 {
		delete(s.m, key)
	}
	return out, true, nil
}

// LRem 按 count 值移除等于 value 的元素，count 为正从头移除、为负从尾移除、为零移除全部
func (m *Memory) LRem(key string, count int, value string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeList {
		return 0, ErrWrongType
	}
	remaining, removed := lrem(e.list, value, count)
	e.list = remaining
	if len(e.list) == 0 {
		delete(s.m, key)
	}
	return int64(removed), nil
}

// LSet 将列表指定下标的元素替换为新值，下标越界返回 ErrIndexRange
func (m *Memory) LSet(key string, index int, value string) error {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return ErrNoKey
	}
	if e.typ != TypeList {
		return ErrWrongType
	}
	i := normalizeIndex(index, len(e.list))
	if i < 0 || i >= len(e.list) || len(e.list) == 0 {
		return ErrIndexRange
	}
	e.list[i] = value
	return nil
}

// LIndex 返回列表指定下标的元素，下标越界返回 false
func (m *Memory) LIndex(key string, index int) (string, bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return "", false, nil
	}
	if e.typ != TypeList {
		s.mu.RUnlock()
		return "", false, ErrWrongType
	}
	i := normalizeIndex(index, len(e.list))
	if i < 0 || i >= len(e.list) {
		s.mu.RUnlock()
		return "", false, nil
	}
	v := e.list[i]
	s.mu.RUnlock()
	return v, true, nil
}

// LInsert 在基准元素前后插入新值，返回插入后的列表长度，基准不存在返回 -1
func (m *Memory) LInsert(key string, before bool, pivot, value string) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeList {
		return 0, ErrWrongType
	}
	idx := indexOf(e.list, pivot)
	if idx < 0 {
		return -1, nil
	}
	if !before {
		idx++
	}
	e.list = insertAt(e.list, idx, value)
	return int64(len(e.list)), nil
}

// LTrim 裁剪列表只保留指定区间，区间为空则删除键
func (m *Memory) LTrim(key string, start, stop int) error {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil
	}
	if e.typ != TypeList {
		return ErrWrongType
	}
	from, to, ok := rangeIndices(start, stop, len(e.list))
	if !ok {
		e.list = nil
		delete(s.m, key)
		return nil
	}
	e.list = e.list[from:to]
	if len(e.list) == 0 {
		delete(s.m, key)
	}
	return nil
}

// rangeIndices 将 start/stop 归一化为半开区间 [from,to)，ok 为 false 表示空区间
func rangeIndices(start, stop, n int) (from, to int, ok bool) {
	if start < 0 {
		start += n
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop += n
	}
	if start >= n || start > stop {
		return 0, 0, false
	}
	if stop >= n {
		stop = n - 1
	}
	return start, stop + 1, true
}

// normalizeIndex 将可能为负的下标归一化到 [0,n-1]，越界返回 -1 或 n
func normalizeIndex(index, n int) int {
	if index < 0 {
		index += n
	}
	if index < 0 {
		return -1
	}
	if index >= n {
		return n
	}
	return index
}

func indexOf(list []string, v string) int {
	for i, s := range list {
		if s == v {
			return i
		}
	}
	return -1
}

func insertAt(list []string, idx int, v string) []string {
	list = append(list, "")
	copy(list[idx+1:], list[idx:])
	list[idx] = v
	return list
}

// lrem 按 count 语义移除 value，count 为正从头、为负从尾、为零移除全部，返回剩余列表与移除数量
func lrem(list []string, value string, count int) ([]string, int) {
	if count == 0 {
		out := make([]string, 0, len(list))
		removed := 0
		for _, s := range list {
			if s == value {
				removed++
				continue
			}
			out = append(out, s)
		}
		return out, removed
	}
	mark := make([]bool, len(list))
	removed := 0
	if count > 0 {
		for i, s := range list {
			if removed == count {
				break
			}
			if s == value {
				mark[i] = true
				removed++
			}
		}
	} else {
		limit := -count
		for i := len(list) - 1; i >= 0; i-- {
			if removed == limit {
				break
			}
			if list[i] == value {
				mark[i] = true
				removed++
			}
		}
	}
	out := make([]string, 0, len(list)-removed)
	for i, s := range list {
		if mark[i] {
			continue
		}
		out = append(out, s)
	}
	return out, removed
}
