/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-05 09:21:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-05 09:21:33
 * @FilePath: \go-memo\store\zset.go
 * @Description: 有序集合类型存储操作，分数映射加惰性有序切片
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"sort"
	"strconv"
)

// ZPair 有序集合成员与分数对
type ZPair struct {
	Member string
	Score  float64
}

// zsetData 有序集合底层数据，order 为按分数排序的缓存，脏时重建
type zsetData struct {
	scores map[string]float64
	order  []zsetItem
	dirty  bool
}

// zsetItem 排序缓存中的单个成员
type zsetItem struct {
	member string
	score  float64
}

// ensureSorted 在 order 缓存脏时重建并排序，分数升序、同分按成员字典序
func (z *zsetData) ensureSorted() {
	if !z.dirty {
		return
	}
	z.order = make([]zsetItem, 0, len(z.scores))
	for m, sc := range z.scores {
		z.order = append(z.order, zsetItem{m, sc})
	}
	sort.Slice(z.order, func(i, j int) bool {
		if z.order[i].score != z.order[j].score {
			return z.order[i].score < z.order[j].score
		}
		return z.order[i].member < z.order[j].member
	})
	z.dirty = false
}

// ZAdd 添加或更新成员分数，返回新增成员数
func (m *Memory) ZAdd(key string, pairs ...ZPair) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeZSet, zset: &zsetData{scores: make(map[string]float64, len(pairs))}}
		s.m[key] = e
	}
	added := 0
	for _, p := range pairs {
		if _, ok := e.zset.scores[p.Member]; !ok {
			added++
		}
		e.zset.scores[p.Member] = p.Score
	}
	if len(pairs) > 0 {
		e.zset.dirty = true
	}
	return added, nil
}

// ZCard 返回成员数量，键不存在返回 0
func (m *Memory) ZCard(key string) (int64, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return 0, nil
	}
	if e.typ != TypeZSet {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}
	n := int64(len(e.zset.scores))
	s.mu.RUnlock()
	return n, nil
}

// ZScore 返回成员分数，成员不存在第二个返回值为 false
func (m *Memory) ZScore(key, member string) (float64, bool, error) {
	s := m.shardOf(key)
	now := unixMilli()
	s.mu.RLock()
	e := s.m[key]
	if e == nil || e.expired(now) {
		s.mu.RUnlock()
		return 0, false, nil
	}
	if e.typ != TypeZSet {
		s.mu.RUnlock()
		return 0, false, ErrWrongType
	}
	sc, ok := e.zset.scores[member]
	s.mu.RUnlock()
	return sc, ok, nil
}

// ZIncrBy 将成员分数递增 delta，返回递增后的分数
func (m *Memory) ZIncrBy(key, member string, delta float64) (float64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e != nil && e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	if e == nil {
		e = &entry{typ: TypeZSet, zset: &zsetData{scores: map[string]float64{}}}
		s.m[key] = e
	}
	next := e.zset.scores[member] + delta
	e.zset.scores[member] = next
	e.zset.dirty = true
	return next, nil
}

// ZRem 移除成员，返回实际移除的成员数
func (m *Memory) ZRem(key string, members ...string) (int, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	removed := 0
	for _, mb := range members {
		if _, ok := e.zset.scores[mb]; ok {
			delete(e.zset.scores, mb)
			removed++
		}
	}
	if removed > 0 {
		e.zset.dirty = true
	}
	if len(e.zset.scores) == 0 {
		delete(s.m, key)
	}
	return removed, nil
}

// ZRange 返回分数升序区间内的成员，start/stop 支持负下标
func (m *Memory) ZRange(key string, start, stop int) ([]ZPair, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, nil
	}
	if e.typ != TypeZSet {
		return nil, ErrWrongType
	}
	e.zset.ensureSorted()
	from, to, ok := rangeIndices(start, stop, len(e.zset.order))
	if !ok {
		return nil, nil
	}
	out := make([]ZPair, 0, to-from)
	for _, it := range e.zset.order[from:to] {
		out = append(out, ZPair{Member: it.member, Score: it.score})
	}
	return out, nil
}

// ZRevRange 返回分数降序区间内的成员，start/stop 从最大分数起算
func (m *Memory) ZRevRange(key string, start, stop int) ([]ZPair, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, nil
	}
	if e.typ != TypeZSet {
		return nil, ErrWrongType
	}
	e.zset.ensureSorted()
	from, to, ok := rangeIndices(start, stop, len(e.zset.order))
	if !ok {
		return nil, nil
	}
	out := make([]ZPair, 0, to-from)
	for i := to - 1; i >= from; i-- {
		it := e.zset.order[i]
		out = append(out, ZPair{Member: it.member, Score: it.score})
	}
	return out, nil
}

// ZRangeByScore 返回分数区间内的成员并按分数升序，支持 LIMIT offset count
func (m *Memory) ZRangeByScore(key, min, max string, offset, count int) ([]ZPair, error) {
	lo, err := parseBound(min)
	if err != nil {
		return nil, ErrNotFloat
	}
	hi, err := parseBound(max)
	if err != nil {
		return nil, ErrNotFloat
	}
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, nil
	}
	if e.typ != TypeZSet {
		return nil, ErrWrongType
	}
	e.zset.ensureSorted()
	var out []ZPair
	for _, it := range e.zset.order {
		if !inScoreRange(lo, hi, it.score) {
			continue
		}
		out = append(out, ZPair{Member: it.member, Score: it.score})
	}
	return applyLimit(out, offset, count), nil
}

// ZRevRangeByScore 返回分数区间内的成员并按分数降序，参数顺序为 max 在前
func (m *Memory) ZRevRangeByScore(key, max, min string, offset, count int) ([]ZPair, error) {
	hi, err := parseBound(max)
	if err != nil {
		return nil, ErrNotFloat
	}
	lo, err := parseBound(min)
	if err != nil {
		return nil, ErrNotFloat
	}
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return nil, nil
	}
	if e.typ != TypeZSet {
		return nil, ErrWrongType
	}
	e.zset.ensureSorted()
	var out []ZPair
	for i := len(e.zset.order) - 1; i >= 0; i-- {
		it := e.zset.order[i]
		if !inScoreRange(lo, hi, it.score) {
			continue
		}
		out = append(out, ZPair{Member: it.member, Score: it.score})
	}
	return applyLimit(out, offset, count), nil
}

// ZRank 返回成员分数升序排名（0 起），成员不存在返回 false
func (m *Memory) ZRank(key, member string) (int64, bool, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, false, nil
	}
	if e.typ != TypeZSet {
		return 0, false, ErrWrongType
	}
	if _, ok := e.zset.scores[member]; !ok {
		return 0, false, nil
	}
	e.zset.ensureSorted()
	for i, it := range e.zset.order {
		if it.member == member {
			return int64(i), true, nil
		}
	}
	return 0, false, nil
}

// ZRevRank 返回成员分数降序排名（0 起），成员不存在返回 false
func (m *Memory) ZRevRank(key, member string) (int64, bool, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, false, nil
	}
	if e.typ != TypeZSet {
		return 0, false, ErrWrongType
	}
	if _, ok := e.zset.scores[member]; !ok {
		return 0, false, nil
	}
	e.zset.ensureSorted()
	n := len(e.zset.order)
	for i, it := range e.zset.order {
		if it.member == member {
			return int64(n - 1 - i), true, nil
		}
	}
	return 0, false, nil
}

// ZCount 返回分数区间内的成员数量
func (m *Memory) ZCount(key, min, max string) (int64, error) {
	lo, err := parseBound(min)
	if err != nil {
		return 0, ErrNotFloat
	}
	hi, err := parseBound(max)
	if err != nil {
		return 0, ErrNotFloat
	}
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	e.zset.ensureSorted()
	n := int64(0)
	for _, it := range e.zset.order {
		if inScoreRange(lo, hi, it.score) {
			n++
		}
	}
	return n, nil
}

// ZRemRangeByRank 移除排名区间内的成员，返回移除数量
func (m *Memory) ZRemRangeByRank(key string, start, stop int) (int64, error) {
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	e.zset.ensureSorted()
	from, to, ok := rangeIndices(start, stop, len(e.zset.order))
	if !ok {
		return 0, nil
	}
	removed := 0
	for _, it := range e.zset.order[from:to] {
		delete(e.zset.scores, it.member)
		removed++
	}
	if removed > 0 {
		e.zset.dirty = true
	}
	if len(e.zset.scores) == 0 {
		delete(s.m, key)
	}
	return int64(removed), nil
}

// ZRemRangeByScore 移除分数区间内的成员，返回移除数量
func (m *Memory) ZRemRangeByScore(key, min, max string) (int64, error) {
	lo, err := parseBound(min)
	if err != nil {
		return 0, ErrNotFloat
	}
	hi, err := parseBound(max)
	if err != nil {
		return 0, ErrNotFloat
	}
	e, s := m.lockEntry(key)
	defer s.mu.Unlock()
	if e == nil {
		return 0, nil
	}
	if e.typ != TypeZSet {
		return 0, ErrWrongType
	}
	e.zset.ensureSorted()
	removed := 0
	for _, it := range e.zset.order {
		if !inScoreRange(lo, hi, it.score) {
			continue
		}
		delete(e.zset.scores, it.member)
		removed++
	}
	if removed > 0 {
		e.zset.dirty = true
	}
	if len(e.zset.scores) == 0 {
		delete(s.m, key)
	}
	return int64(removed), nil
}

// bound 分数区间边界，negInf/posInf 表示无穷，exclusive 表示开区间
type bound struct {
	score     float64
	exclusive bool
	negInf    bool
	posInf    bool
}

// parseBound 解析区间边界，支持 (score 开区间与 -inf/+inf 无穷
func parseBound(s string) (bound, error) {
	switch s {
	case "-inf":
		return bound{negInf: true}, nil
	case "+inf", "inf":
		return bound{posInf: true}, nil
	}
	if len(s) > 0 && s[0] == '(' {
		f, err := strconv.ParseFloat(s[1:], 64)
		if err != nil {
			return bound{}, err
		}
		return bound{score: f, exclusive: true}, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return bound{}, err
	}
	return bound{score: f}, nil
}

// inScoreRange 判断分数是否落在 [lo, hi] 区间内
func inScoreRange(lo, hi bound, score float64) bool {
	if !lo.negInf {
		if lo.exclusive {
			if score <= lo.score {
				return false
			}
		} else if score < lo.score {
			return false
		}
	}
	if !hi.posInf {
		if hi.exclusive {
			if score >= hi.score {
				return false
			}
		} else if score > hi.score {
			return false
		}
	}
	return true
}

// applyLimit 对成员结果应用 LIMIT offset count，count 非正表示不限制
func applyLimit(pairs []ZPair, offset, count int) []ZPair {
	if offset > 0 {
		if offset >= len(pairs) {
			return nil
		}
		pairs = pairs[offset:]
	}
	if count > 0 && count < len(pairs) {
		pairs = pairs[:count]
	}
	return pairs
}
