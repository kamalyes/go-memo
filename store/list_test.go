/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-12 09:31:20
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-12 09:31:20
 * @FilePath: \go-memo\store\list_test.go
 * @Description: 列表类型存储单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"reflect"
	"testing"
)

func TestListPushOrder(t *testing.T) {
	m := New()
	if _, err := m.RPush("l", "a", "b", "c"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.LPush("l", "x", "y"); err != nil {
		t.Fatal(err)
	}
	got, err := m.LRange("l", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"y", "x", "a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LRange = %v, want %v", got, want)
	}
	if n, _ := m.LLen("l"); n != 5 {
		t.Fatalf("LLen = %d, want 5", n)
	}
}

func TestListPushXMissing(t *testing.T) {
	m := New()
	if n, err := m.LPushX("l", "a"); err != nil || n != 0 {
		t.Fatalf("LPushX = %d, %v; want 0, nil", n, err)
	}
	if n, err := m.RPushX("l", "a"); err != nil || n != 0 {
		t.Fatalf("RPushX = %d, %v; want 0, nil", n, err)
	}
	if m.Type("l") != TypeNone {
		t.Fatal("pushx should not create missing key")
	}
}

func TestListRangeNegative(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "b", "c", "d")
	got, err := m.LRange("l", -2, -1)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"c", "d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LRange = %v, want %v", got, want)
	}
	if got, _ := m.LRange("l", 3, 1); got != nil {
		t.Fatalf("LRange empty range = %v, want nil", got)
	}
}

func TestListPop(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "b", "c")
	vals, existed, err := m.LPop("l", 2)
	if err != nil || !existed {
		t.Fatalf("LPop = %v, %v, %v", vals, existed, err)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(vals, want) {
		t.Fatalf("LPop = %v, want %v", vals, want)
	}
	vals, existed, err = m.RPop("l", 5)
	if err != nil || !existed {
		t.Fatalf("RPop = %v, %v, %v", vals, existed, err)
	}
	if want := []string{"c"}; !reflect.DeepEqual(vals, want) {
		t.Fatalf("RPop = %v, want %v", vals, want)
	}
	if m.Type("l") != TypeNone {
		t.Fatal("pop to empty should delete key")
	}
}

func TestListRem(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "b", "a", "c", "a")
	if n, _ := m.LRem("l", 1, "a"); n != 1 {
		t.Fatalf("LRem head = %d, want 1", n)
	}
	if got, _ := m.LRange("l", 0, -1); !reflect.DeepEqual(got, []string{"b", "a", "c", "a"}) {
		t.Fatalf("after LRem head = %v", got)
	}
	if n, _ := m.LRem("l", 0, "a"); n != 2 {
		t.Fatalf("LRem all = %d, want 2", n)
	}
	if got, _ := m.LRange("l", 0, -1); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("after LRem all = %v", got)
	}

	m2 := New()
	_, _ = m2.RPush("l", "a", "b", "a", "c", "a")
	if n, _ := m2.LRem("l", -1, "a"); n != 1 {
		t.Fatalf("LRem tail = %d, want 1", n)
	}
	if got, _ := m2.LRange("l", 0, -1); !reflect.DeepEqual(got, []string{"a", "b", "a", "c"}) {
		t.Fatalf("after LRem tail = %v", got)
	}
}

func TestListSetIndex(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "b", "c")
	if err := m.LSet("l", 1, "B"); err != nil {
		t.Fatal(err)
	}
	if v, ok, _ := m.LIndex("l", 1); !ok || v != "B" {
		t.Fatalf("LIndex = %q, %v; want B, true", v, ok)
	}
	if v, ok, _ := m.LIndex("l", -1); !ok || v != "c" {
		t.Fatalf("LIndex -1 = %q, %v; want c, true", v, ok)
	}
	if err := m.LSet("l", 10, "x"); !errors.Is(err, ErrIndexRange) {
		t.Fatalf("LSet out of range = %v, want ErrIndexRange", err)
	}
	if err := m.LSet("missing", 0, "x"); !errors.Is(err, ErrNoKey) {
		t.Fatalf("LSet missing = %v, want ErrNoKey", err)
	}
}

func TestListInsert(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "c")
	if n, err := m.LInsert("l", true, "c", "b"); err != nil || n != 3 {
		t.Fatalf("LInsert before = %d, %v; want 3, nil", n, err)
	}
	if got, _ := m.LRange("l", 0, -1); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("after insert before = %v", got)
	}
	if n, _ := m.LInsert("l", false, "a", "x"); n != 4 {
		t.Fatalf("LInsert after = %d, want 4", n)
	}
	if n, _ := m.LInsert("l", true, "missing", "x"); n != -1 {
		t.Fatalf("LInsert missing pivot = %d, want -1", n)
	}
}

func TestListTrim(t *testing.T) {
	m := New()
	_, _ = m.RPush("l", "a", "b", "c", "d")
	if err := m.LTrim("l", 1, 2); err != nil {
		t.Fatal(err)
	}
	if got, _ := m.LRange("l", 0, -1); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("after trim = %v", got)
	}
}

func TestListWrongType(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if _, err := m.LPush("k", "a"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("LPush wrong type = %v, want ErrWrongType", err)
	}
	if _, err := m.LLen("k"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("LLen wrong type = %v, want ErrWrongType", err)
	}
}