/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-13 09:18:27
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-13 09:18:27
 * @FilePath: \go-memo\store\set_test.go
 * @Description: 集合类型存储单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

import (
	"errors"
	"testing"
)

func contains(vals []string, v string) bool {
	for _, s := range vals {
		if s == v {
			return true
		}
	}
	return false
}

func TestSetAddRem(t *testing.T) {
	m := New()
	if n, err := m.SAdd("s", "a", "b", "a"); err != nil || n != 2 {
		t.Fatalf("SAdd = %d, %v; want 2, nil", n, err)
	}
	if n, _ := m.SAdd("s", "b", "c"); n != 1 {
		t.Fatalf("SAdd = %d, want 1", n)
	}
	if n, _ := m.SCard("s"); n != 3 {
		t.Fatalf("SCard = %d, want 3", n)
	}
	if ok, _ := m.SIsMember("s", "a"); !ok {
		t.Fatal("SIsMember a should be true")
	}
	if n, _ := m.SRem("s", "a", "z"); n != 1 {
		t.Fatalf("SRem = %d, want 1", n)
	}
	if members, _ := m.SMembers("s"); len(members) != 2 {
		t.Fatalf("SMembers len = %d, want 2", len(members))
	}
}

func TestSetPop(t *testing.T) {
	m := New()
	_, _ = m.SAdd("s", "a", "b", "c")
	vals, existed, err := m.SPop("s", 2)
	if err != nil || !existed || len(vals) != 2 {
		t.Fatalf("SPop = %v, %v, %v; want 2 members", vals, existed, err)
	}
	if n, _ := m.SCard("s"); n != 1 {
		t.Fatalf("SCard after pop = %d, want 1", n)
	}
	vals, existed, _ = m.SPop("s", 5)
	if !existed || len(vals) != 1 {
		t.Fatalf("SPop all = %v, want 1 member", vals)
	}
	if m.Type("s") != TypeNone {
		t.Fatal("pop to empty should delete key")
	}
}

func TestSetRandMember(t *testing.T) {
	m := New()
	_, _ = m.SAdd("s", "a", "b", "c")
	if vals, _ := m.SRandMember("s", 2); len(vals) != 2 {
		t.Fatalf("SRandMember = %v, want 2", vals)
	}
	if vals, _ := m.SRandMember("s", -5); len(vals) != 5 {
		t.Fatalf("SRandMember neg = %v, want 5", vals)
	}
	if vals, _ := m.SRandMember("s", 0); len(vals) != 1 {
		t.Fatalf("SRandMember zero = %v, want 1", vals)
	}
}

func TestSetMove(t *testing.T) {
	m := New()
	_, _ = m.SAdd("s1", "a", "b")
	_, _ = m.SAdd("s2", "c")
	if ok, _ := m.SMove("s1", "s2", "a"); !ok {
		t.Fatal("SMove should be true")
	}
	if ok, _ := m.SIsMember("s1", "a"); ok {
		t.Fatal("s1 should not contain a")
	}
	if ok, _ := m.SIsMember("s2", "a"); !ok {
		t.Fatal("s2 should contain a")
	}
	if ok, _ := m.SMove("s1", "s2", "missing"); ok {
		t.Fatal("SMove missing should be false")
	}
}

func TestSetAlgebra(t *testing.T) {
	m := New()
	_, _ = m.SAdd("a", "1", "2", "3")
	_, _ = m.SAdd("b", "2", "3", "5")
	inter, _ := m.SInter("a", "b")
	if len(inter) != 2 || !contains(inter, "2") || !contains(inter, "3") {
		t.Fatalf("SInter = %v", inter)
	}
	union, _ := m.SUnion("a", "b")
	if len(union) != 4 {
		t.Fatalf("SUnion = %v, want 4", union)
	}
	diff, _ := m.SDiff("a", "b")
	if len(diff) != 1 || !contains(diff, "1") {
		t.Fatalf("SDiff = %v, want [1]", diff)
	}
	if n, _ := m.SInterStore("dst", "a", "b"); n != 2 {
		t.Fatalf("SInterStore = %d, want 2", n)
	}
	if n, _ := m.SCard("dst"); n != 2 {
		t.Fatalf("SCard dst = %d, want 2", n)
	}
}

func TestSetWrongType(t *testing.T) {
	m := New()
	m.Set("k", "v")
	if _, err := m.SAdd("k", "a"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("SAdd wrong type = %v, want ErrWrongType", err)
	}
}
