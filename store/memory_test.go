/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 11:15:28
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-17 11:15:28
 * @FilePath: \go-memo\store\memory_test.go
 * @Description: 分片内存键值存储单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"testing"
)

func TestSetGet(t *testing.T) {
	m := New()
	m.Set("foo", "bar")
	v, ok := m.Get("foo")
	if !ok || v != "bar" {
		t.Fatalf("Get = %q, %v; want bar, true", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	m := New()
	if _, ok := m.Get("missing"); ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestIncrByNewKey(t *testing.T) {
	m := New()
	n, err := m.IncrBy("cnt", 10)
	if err != nil {
		t.Fatalf("IncrBy: %v", err)
	}
	if n != 10 {
		t.Fatalf("IncrBy = %d, want 10", n)
	}
	n, err = m.IncrBy("cnt", -3)
	if err != nil {
		t.Fatalf("IncrBy: %v", err)
	}
	if n != 7 {
		t.Fatalf("IncrBy = %d, want 7", n)
	}
}

func TestIncrByNotInteger(t *testing.T) {
	m := New()
	m.Set("str", "abc")
	if _, err := m.IncrBy("str", 1); !errors.Is(err, ErrNotInteger) {
		t.Fatalf("expected errNotInteger, got %v", err)
	}
}

func TestDel(t *testing.T) {
	m := New()
	m.Set("a", "1")
	m.Set("b", "2")
	m.Set("c", "3")
	if n := m.Del("a", "b", "missing"); n != 2 {
		t.Fatalf("Del count = %d, want 2", n)
	}
	if _, ok := m.Get("a"); ok {
		t.Fatal("key a should be deleted")
	}
	if _, ok := m.Get("c"); !ok {
		t.Fatal("key c should remain")
	}
}

func TestExpire(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if !m.Expire("k", unixMilli()+60_000) {
		t.Fatal("Expire should succeed on live key")
	}
	ttl := m.TTL("k")
	if ttl <= 0 || ttl > 60_000 {
		t.Fatalf("TTL = %d, want in (0, 60000]", ttl)
	}
}

func TestExpireImmediatelyExpired(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if !m.Expire("k", unixMilli()-1) {
		t.Fatal("Expire should succeed on live key")
	}
	if got := m.TTL("k"); got != TTLNotExist {
		t.Fatalf("TTL = %d, want %d", got, TTLNotExist)
	}
	if _, ok := m.Get("k"); ok {
		t.Fatal("expired key should not be readable")
	}
}

func TestExpireMissingKey(t *testing.T) {
	m := New()
	if m.Expire("missing", unixMilli()+1000) {
		t.Fatal("Expire should fail on missing key")
	}
}

func TestTTLSemantics(t *testing.T) {
	m := New()
	if got := m.TTL("missing"); got != TTLNotExist {
		t.Fatalf("missing TTL = %d, want %d", got, TTLNotExist)
	}
	m.Set("forever", "v")
	if got := m.TTL("forever"); got != TTLNoExpire {
		t.Fatalf("no-expire TTL = %d, want %d", got, TTLNoExpire)
	}
}

func TestSetClearsExpiry(t *testing.T) {
	m := New()
	m.Set("k", "v")
	m.Expire("k", unixMilli()+60_000)
	m.Set("k", "v2")
	if got := m.TTL("k"); got != TTLNoExpire {
		t.Fatalf("TTL = %d, want %d after Set", got, TTLNoExpire)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	m := New()
	done := make(chan struct{})
	for i := 0; i < 32; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 1000; j++ {
				m.Set("shared", "v")
				m.IncrBy("counter", 1)
				m.Get("shared")
			}
		}()
	}
	for i := 0; i < 32; i++ {
		<-done
	}
	n, err := m.IncrBy("counter", 0)
	if err != nil {
		t.Fatalf("IncrBy: %v", err)
	}
	if n != 32*1000 {
		t.Fatalf("counter = %d, want %d", n, 32*1000)
	}
}

func TestKeys(t *testing.T) {
	m := New()
	m.Set("a", "1")
	m.Set("b", "2")
	m.Expire("a", unixMilli()-1)
	keys := m.Keys()
	if len(keys) != 1 || keys[0] != "b" {
		t.Fatalf("Keys = %v, want [b]", keys)
	}
}

func TestExists(t *testing.T) {
	m := New()
	m.Set("a", "1")
	if !m.Exists("a") {
		t.Fatal("Exists a = false, want true")
	}
	if m.Exists("missing") {
		t.Fatal("Exists missing = true, want false")
	}
}

func TestKeyspace(t *testing.T) {
	m := New()
	m.Set("a", "1")
	m.Set("b", "2")
	m.Expire("b", unixMilli()+60_000)
	keys, expires := m.Keyspace()
	if keys != 2 || expires != 1 {
		t.Fatalf("Keyspace = (%d,%d), want (2,1)", keys, expires)
	}
}

func TestUsedMemory(t *testing.T) {
	m := New()
	if got := m.UsedMemory(); got != 0 {
		t.Fatalf("empty UsedMemory = %d, want 0", got)
	}
	m.Set("foo", "bar")
	if got := m.UsedMemory(); got <= 0 {
		t.Fatalf("UsedMemory = %d, want > 0", got)
	}
	m.Expire("foo", unixMilli()-1)
	if got := m.UsedMemory(); got != 0 {
		t.Fatalf("expired UsedMemory = %d, want 0", got)
	}
}
