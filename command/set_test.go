/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-16 09:16:38
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-16 09:16:38
 * @FilePath: \go-memo\command\set_test.go
 * @Description: 集合类型命令单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"reflect"
	"testing"

	"github.com/kamalyes/go-memo/store"
)

func TestSetCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantInt(t, execCmd(t, r, st, "SADD", "s", "a", "b", "a"), 2)
	wantInt(t, execCmd(t, r, st, "SADD", "s", "c"), 1)
	wantInt(t, execCmd(t, r, st, "SCARD", "s"), 3)
	wantInt(t, execCmd(t, r, st, "SISMEMBER", "s", "a"), 1)
	wantInt(t, execCmd(t, r, st, "SISMEMBER", "s", "z"), 0)
	if got := asStrs(t, execCmd(t, r, st, "SMEMBERS", "s")); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("SMEMBERS = %v", got)
	}
	wantInt(t, execCmd(t, r, st, "SREM", "s", "a"), 1)
}

func TestSetMoveAndPop(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SADD", "s1", "a", "b")
	execCmd(t, r, st, "SADD", "s2", "c")
	wantInt(t, execCmd(t, r, st, "SMOVE", "s1", "s2", "a"), 1)
	wantInt(t, execCmd(t, r, st, "SISMEMBER", "s1", "a"), 0)
	wantInt(t, execCmd(t, r, st, "SISMEMBER", "s2", "a"), 1)
	wantBulk(t, execCmd(t, r, st, "SPOP", "s1"), "b")
}

func TestSetAlgebraCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SADD", "a", "1", "2", "3")
	execCmd(t, r, st, "SADD", "b", "2", "3", "5")
	if got := asStrs(t, execCmd(t, r, st, "SINTER", "a", "b")); !reflect.DeepEqual(got, []string{"2", "3"}) {
		t.Fatalf("SINTER = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "SUNION", "a", "b")); !reflect.DeepEqual(got, []string{"1", "2", "3", "5"}) {
		t.Fatalf("SUNION = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "SDIFF", "a", "b")); !reflect.DeepEqual(got, []string{"1"}) {
		t.Fatalf("SDIFF = %v", got)
	}
	wantInt(t, execCmd(t, r, st, "SINTERSTORE", "dst", "a", "b"), 2)
	wantInt(t, execCmd(t, r, st, "SCARD", "dst"), 2)
	wantInt(t, execCmd(t, r, st, "SUNIONSTORE", "u", "a", "b"), 4)
	wantInt(t, execCmd(t, r, st, "SDIFFSTORE", "d", "a", "b"), 1)
}

func TestSetCommandErrors(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SET", "k", "v")
	wantErr(t, execCmd(t, r, st, "SADD", "k", "a"), store.ErrWrongType.Error())
	wantErr(t, execCmd(t, r, st, "SMEMBERS"), errWrongArgs("smembers"))
}