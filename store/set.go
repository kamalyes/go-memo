/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-03 09:16:20
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-03 09:16:20
 * @FilePath: \go-memo\store\set.go
 * @Description: 集合类型存储操作，底层为去重映射
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// SAdd 添加成员，返回新增成员数
func (m *Memory) SAdd(key string, members ...string) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeSet {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeSet, set: make(map[string]struct{}, len(members))}
		s.m[key] = e
	}
	added := 0
	for _, mb := range members {
		if _, ok := e.set[mb]; !ok {
			e.set[mb] = struct{}{}
			added++
		}
	}
	return added, nil
}

// SRem 移除成员，返回实际移除的成员数
func (m *Memory) SRem(key string, members ...string) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeSet {
		return 0, ErrWrongType
	}
	removed := 0
	for _, mb := range members {
		if _, ok := e.set[mb]; ok {
			delete(e.set, mb)
			removed++
		}
	}
	if len(e.set) == 0 {
		delete(s.m, key)
	}
	return removed, nil
}

// SMembers 返回全部成员副本
func (m *Memory) SMembers(key string) ([]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, nil
	}
	if e.typ != TypeSet {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	out := make([]string, 0, len(e.set))
	for v := range e.set {
		out = append(out, v)
	}
	s.mu.RUnlock()
	return out, nil
}

// SIsMember 判断成员是否存在
func (m *Memory) SIsMember(key, member string) (bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return false, nil
	}
	if e.typ != TypeSet {
		s.mu.RUnlock()
		return false, ErrWrongType
	}
	_, ok := e.set[member]
	s.mu.RUnlock()
	return ok, nil
}

// SCard 返回成员数量，键不存在返回 0
func (m *Memory) SCard(key string) (int64, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return 0, nil
	}
	if e.typ != TypeSet {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}
	n := int64(len(e.set))
	s.mu.RUnlock()
	return n, nil
}

// SPop 随机移除并返回 count 个成员，existed 标识键是否存在且为集合
func (m *Memory) SPop(key string, count int) ([]string, bool, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, false, nil
	}
	if e.typ != TypeSet {
		return nil, false, ErrWrongType
	}
	if count <= 0 {
		count = 1
	}
	if count > len(e.set) {
		count = len(e.set)
	}
	out := make([]string, 0, count)
	for mb := range e.set {
		if len(out) == count {
			break
		}
		out = append(out, mb)
		delete(e.set, mb)
	}
	if len(e.set) == 0 {
		delete(s.m, key)
	}
	return out, true, nil
}

// SRandMember 随机返回成员，count 为正不重复、为负可重复
func (m *Memory) SRandMember(key string, count int) ([]string, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, nil
	}
	if e.typ != TypeSet {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}
	all := make([]string, 0, len(e.set))
	for v := range e.set {
		all = append(all, v)
	}
	s.mu.RUnlock()
	return sampleMembers(all, count), nil
}

// sampleMembers 从成员列表采样，count 为正不重复、为负循环可重复、为零取一个
func sampleMembers(all []string, count int) []string {
	if len(all) == 0 {
		return nil
	}
	if count == 0 {
		count = 1
	}
	if count < 0 {
		n := -count
		out := make([]string, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, all[i%len(all)])
		}
		return out
	}
	if count > len(all) {
		count = len(all)
	}
	return all[:count]
}

// SMove 将成员从源集合移动到目标集合，返回是否移动成功
func (m *Memory) SMove(src, dst, member string) (bool, error) {
	si := fnv32(src) % shardCount
	di := fnv32(dst) % shardCount
	if si == di {
		s := m.shards[si]
		s.mu.Lock()
		defer s.mu.Unlock()
		return smovePair(s, s, src, dst, member)
	}
	lo, hi := si, di
	if lo > hi {
		lo, hi = hi, lo
	}
	slo, shi := m.shards[lo], m.shards[hi]
	slo.mu.Lock()
	shi.mu.Lock()
	defer shi.mu.Unlock()
	defer slo.mu.Unlock()
	return smovePair(m.shards[si], m.shards[di], src, dst, member)
}

// smovePair 在分片已加锁前提下移动成员，目标类型不匹配返回 ErrWrongType
func smovePair(srcS, dstS *shard, src, dst, member string) (bool, error) {
	now := unixMilli()
	se := srcS.m[src]
	if se == nil || se.expired(now) {
		if se != nil {
			delete(srcS.m, src)
		}
		return false, nil
	}
	if se.typ != TypeSet {
		return false, ErrWrongType
	}
	if _, ok := se.set[member]; !ok {
		return false, nil
	}
	de := dstS.m[dst]
	if de != nil && de.expired(now) {
		delete(dstS.m, dst)
		de = nil
	}
	if de != nil && de.typ != TypeSet {
		return false, ErrWrongType
	}
	delete(se.set, member)
	if len(se.set) == 0 {
		delete(srcS.m, src)
	}
	if de == nil {
		de = &entry{typ: TypeSet, set: map[string]struct{}{member: {}}}
		dstS.m[dst] = de
	} else {
		de.set[member] = struct{}{}
	}
	return true, nil
}

// SInter 返回多个集合的交集
func (m *Memory) SInter(keys ...string) ([]string, error) {
	sets, err := m.collectSets(keys, true)
	if err != nil || sets == nil {
		return nil, err
	}
	min := sets[0]
	for _, s := range sets {
		if len(s) < len(min) {
			min = s
		}
	}
	out := make([]string, 0, len(min))
	for v := range min {
		all := true
		for _, s := range sets {
			if _, ok := s[v]; !ok {
				all = false
				break
			}
		}
		if all {
			out = append(out, v)
		}
	}
	return out, nil
}

// SUnion 返回多个集合的并集
func (m *Memory) SUnion(keys ...string) ([]string, error) {
	sets, err := m.collectSets(keys, false)
	if err != nil {
		return nil, err
	}
	merged := map[string]struct{}{}
	for _, s := range sets {
		for v := range s {
			merged[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(merged))
	for v := range merged {
		out = append(out, v)
	}
	return out, nil
}

// SDiff 返回第一个集合相对其余集合的差集
func (m *Memory) SDiff(keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	first, ok, err := m.setSnapshot(keys[0])
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	exclude := map[string]struct{}{}
	for _, k := range keys[1:] {
		s, ok2, err2 := m.setSnapshot(k)
		if err2 != nil {
			return nil, err2
		}
		if !ok2 {
			continue
		}
		for v := range s {
			exclude[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(first))
	for v := range first {
		if _, in := exclude[v]; !in {
			out = append(out, v)
		}
	}
	return out, nil
}

// SInterStore 计算交集并写入目标键
func (m *Memory) SInterStore(dst string, keys ...string) (int64, error) {
	result, err := m.SInter(keys...)
	if err != nil {
		return 0, err
	}
	return m.storeSetResult(dst, result)
}

// SUnionStore 计算并集并写入目标键
func (m *Memory) SUnionStore(dst string, keys ...string) (int64, error) {
	result, err := m.SUnion(keys...)
	if err != nil {
		return 0, err
	}
	return m.storeSetResult(dst, result)
}

// SDiffStore 计算差集并写入目标键
func (m *Memory) SDiffStore(dst string, keys ...string) (int64, error) {
	result, err := m.SDiff(keys...)
	if err != nil {
		return 0, err
	}
	return m.storeSetResult(dst, result)
}

// storeSetResult 将集合结果写入目标键，空结果删除键
func (m *Memory) storeSetResult(dst string, result []string) (int64, error) {
	e, s := m.lockEntry(dst)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeSet {
		return 0, ErrWrongType
	}
	if len(result) == 0 {
		if e != nil {
			delete(s.m, dst)
		}
		return 0, nil
	}
	if e == nil {
		e = &entry{typ: TypeSet}
		s.m[dst] = e
	}
	e.set = make(map[string]struct{}, len(result))
	for _, v := range result {
		e.set[v] = struct{}{}
	}
	return int64(len(result)), nil
}

// setSnapshot 读锁采集集合成员副本，返回成员映射、键是否存在与错误
func (m *Memory) setSnapshot(key string) (map[string]struct{}, bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return nil, false, nil
	}
	if e.typ != TypeSet {
		s.mu.RUnlock()
		return nil, false, ErrWrongType
	}
	out := make(map[string]struct{}, len(e.set))
	for v := range e.set {
		out[v] = struct{}{}
	}
	s.mu.RUnlock()
	return out, true, nil
}

// collectSets 采集多个集合成员副本，missingIsEmpty 为真时任一键缺失返回空集合
func (m *Memory) collectSets(keys []string, missingIsEmpty bool) ([]map[string]struct{}, error) {
	sets := make([]map[string]struct{}, 0, len(keys))
	for _, k := range keys {
		s, ok, err := m.setSnapshot(k)
		if err != nil {
			return nil, err
		}
		if !ok {
			if missingIsEmpty {
				return nil, nil
			}
			continue
		}
		sets = append(sets, s)
	}
	return sets, nil
}
