/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-12 09:35:15
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-12 09:35:15
 * @FilePath: \go-memo\store\hash_test.go
 * @Description: 哈希类型存储单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"testing"
)

func TestHashSetGet(t *testing.T) {
	m := New()
	if n, err := m.HSet("h", "f1", "v1"); err != nil || n != 1 {
		t.Fatalf("HSet = %d, %v; want 1, nil", n, err)
	}
	if n, _ := m.HSet("h", "f1", "v2"); n != 0 {
		t.Fatalf("HSet update = %d, want 0", n)
	}
	if n, _ := m.HSet("h", "f2", "v2"); n != 1 {
		t.Fatalf("HSet add = %d, want 1", n)
	}
	if v, ok, _ := m.HGet("h", "f1"); !ok || v != "v2" {
		t.Fatalf("HGet = %q, %v; want v2, true", v, ok)
	}
	if _, ok, _ := m.HGet("h", "missing"); ok {
		t.Fatal("HGet missing should be false")
	}
}

func TestHashSetMap(t *testing.T) {
	m := New()
	if n, err := m.HSetMap("h", map[string]string{"a": "1", "b": "2"}); err != nil || n != 2 {
		t.Fatalf("HSetMap = %d, %v; want 2, nil", n, err)
	}
	if n, _ := m.HSetMap("h", map[string]string{"a": "x", "c": "3"}); n != 1 {
		t.Fatalf("HSetMap update = %d, want 1", n)
	}
	if n, _ := m.HLen("h"); n != 3 {
		t.Fatalf("HLen = %d, want 3", n)
	}
}

func TestHashSetNX(t *testing.T) {
	m := New()
	if ok, _ := m.HSetNX("h", "f", "v"); !ok {
		t.Fatal("HSetNX first should be true")
	}
	if ok, _ := m.HSetNX("h", "f", "v2"); ok {
		t.Fatal("HSetNX existing should be false")
	}
	if v, ok, _ := m.HGet("h", "f"); !ok || v != "v" {
		t.Fatalf("HGet = %q, %v; want v, true", v, ok)
	}
}

func TestHashBulk(t *testing.T) {
	m := New()
	_, _ = m.HSetMap("h", map[string]string{"a": "1", "b": "2", "c": "3"})
	vals, oks, err := m.HMGet("h", "a", "missing", "c")
	if err != nil {
		t.Fatal(err)
	}
	if !oks[0] || vals[0] != "1" || oks[1] || vals[1] != "" || !oks[2] || vals[2] != "3" {
		t.Fatalf("HMGet = %v, %v", vals, oks)
	}
	if ok, _ := m.HExists("h", "b"); !ok {
		t.Fatal("HExists b should be true")
	}
	if all, _ := m.HGetAll("h"); len(all) != 3 {
		t.Fatalf("HGetAll len = %d, want 3", len(all))
	}
}

func TestHashDelAndClear(t *testing.T) {
	m := New()
	_, _ = m.HSetMap("h", map[string]string{"a": "1", "b": "2"})
	if n, _ := m.HDel("h", "a", "missing"); n != 1 {
		t.Fatalf("HDel = %d, want 1", n)
	}
	if n, _ := m.HDel("h", "b"); n != 1 {
		t.Fatalf("HDel = %d, want 1", n)
	}
	if m.Type("h") != TypeNone {
		t.Fatal("empty hash should delete key")
	}
}

func TestHashIncrBy(t *testing.T) {
	m := New()
	if n, err := m.HIncrBy("h", "n", 5); err != nil || n != 5 {
		t.Fatalf("HIncrBy = %d, %v; want 5, nil", n, err)
	}
	if n, _ := m.HIncrBy("h", "n", -2); n != 3 {
		t.Fatalf("HIncrBy = %d, want 3", n)
	}
	_, _ = m.HSet("h", "s", "abc")
	if _, err := m.HIncrBy("h", "s", 1); !errors.Is(err, ErrNotInteger) {
		t.Fatalf("HIncrBy non-int = %v, want ErrNotInteger", err)
	}
}

func TestHashWrongType(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if _, err := m.HSet("k", "f", "v"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("HSet wrong type = %v, want ErrWrongType", err)
	}
	if _, _, err := m.HGet("k", "f"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("HGet wrong type = %v, want ErrWrongType", err)
	}
}