/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-13 09:22:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-13 09:22:11
 * @FilePath: \go-memo\store\zset_test.go
 * @Description: 有序集合类型存储单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"reflect"
	"testing"
)

func members(pairs []ZPair) []string {
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.Member
	}
	return out
}

func TestZSetBasic(t *testing.T) {
	m := New()
	if n, err := m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 2}, ZPair{"c", 1.5}); err != nil || n != 3 {
		t.Fatalf("ZAdd = %d, %v; want 3, nil", n, err)
	}
	if n, _ := m.ZAdd("z", ZPair{"a", 5}); n != 0 {
		t.Fatalf("ZAdd update = %d, want 0", n)
	}
	if n, _ := m.ZCard("z"); n != 3 {
		t.Fatalf("ZCard = %d, want 3", n)
	}
	if sc, ok, _ := m.ZScore("z", "a"); !ok || sc != 5 {
		t.Fatalf("ZScore = %v, %v; want 5, true", sc, ok)
	}
	if _, ok, _ := m.ZScore("z", "missing"); ok {
		t.Fatal("ZScore missing should be false")
	}
}

func TestZSetRange(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 3}, ZPair{"c", 2})
	pairs, err := m.ZRange("z", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "c", "b"}; !reflect.DeepEqual(members(pairs), want) {
		t.Fatalf("ZRange = %v, want %v", members(pairs), want)
	}
	rev, _ := m.ZRevRange("z", 0, -1)
	if want := []string{"b", "c", "a"}; !reflect.DeepEqual(members(rev), want) {
		t.Fatalf("ZRevRange = %v, want %v", members(rev), want)
	}
}

func TestZSetIncrBy(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1})
	if v, err := m.ZIncrBy("z", "a", 2); err != nil || v != 3 {
		t.Fatalf("ZIncrBy = %v, %v; want 3, nil", v, err)
	}
	if v, _ := m.ZIncrBy("z", "b", 2); v != 2 {
		t.Fatalf("ZIncrBy new = %v, want 2", v)
	}
}

func TestZSetRem(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 2})
	if n, _ := m.ZRem("z", "a", "missing"); n != 1 {
		t.Fatalf("ZRem = %d, want 1", n)
	}
	if n, _ := m.ZRem("z", "b"); n != 1 {
		t.Fatalf("ZRem = %d, want 1", n)
	}
	if m.Type("z") != TypeNone {
		t.Fatal("empty zset should delete key")
	}
}

func TestZSetRangeByScore(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 2}, ZPair{"c", 3})
	pairs, err := m.ZRangeByScore("z", "(1", "3", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"b", "c"}; !reflect.DeepEqual(members(pairs), want) {
		t.Fatalf("ZRangeByScore = %v, want %v", members(pairs), want)
	}
	pairs, _ = m.ZRangeByScore("z", "-inf", "+inf", 1, 1)
	if want := []string{"b"}; !reflect.DeepEqual(members(pairs), want) {
		t.Fatalf("ZRangeByScore limit = %v, want %v", members(pairs), want)
	}
	rev, _ := m.ZRevRangeByScore("z", "3", "1", 0, -1)
	if want := []string{"c", "b", "a"}; !reflect.DeepEqual(members(rev), want) {
		t.Fatalf("ZRevRangeByScore = %v, want %v", members(rev), want)
	}
}

func TestZSetRank(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 3}, ZPair{"c", 2})
	if rank, ok, _ := m.ZRank("z", "a"); !ok || rank != 0 {
		t.Fatalf("ZRank a = %d, %v; want 0, true", rank, ok)
	}
	if rank, _, _ := m.ZRank("z", "b"); rank != 2 {
		t.Fatalf("ZRank b = %d, want 2", rank)
	}
	if rank, _, _ := m.ZRevRank("z", "b"); rank != 0 {
		t.Fatalf("ZRevRank b = %d, want 0", rank)
	}
	if _, ok, _ := m.ZRank("z", "missing"); ok {
		t.Fatal("ZRank missing should be false")
	}
}

func TestZSetCountAndRemRange(t *testing.T) {
	m := New()
	_, _ = m.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 2}, ZPair{"c", 3})
	if n, _ := m.ZCount("z", "1", "2"); n != 2 {
		t.Fatalf("ZCount = %d, want 2", n)
	}
	if n, _ := m.ZRemRangeByScore("z", "2", "+inf"); n != 2 {
		t.Fatalf("ZRemRangeByScore = %d, want 2", n)
	}
	if n, _ := m.ZCard("z"); n != 1 {
		t.Fatalf("ZCard after rem = %d, want 1", n)
	}

	m2 := New()
	_, _ = m2.ZAdd("z", ZPair{"a", 1}, ZPair{"b", 2}, ZPair{"c", 3})
	if n, _ := m2.ZRemRangeByRank("z", 0, 0); n != 1 {
		t.Fatalf("ZRemRangeByRank = %d, want 1", n)
	}
	if _, ok, _ := m2.ZScore("z", "a"); ok {
		t.Fatal("a should be removed by rank")
	}
}

func TestZSetWrongType(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if _, err := m.ZAdd("k", ZPair{"a", 1}); !errors.Is(err, ErrWrongType) {
		t.Fatalf("ZAdd wrong type = %v, want ErrWrongType", err)
	}
}