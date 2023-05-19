/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-15 09:15:21
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-15 09:15:21
 * @FilePath: \go-memo\command\list_test.go
 * @Description: 列表类型命令单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"reflect"
	"testing"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func TestListPushRange(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantInt(t, execCmd(t, r, st, "RPUSH", "l", "a", "b", "c"), 3)
	wantInt(t, execCmd(t, r, st, "LPUSH", "l", "x"), 4)
	want := []string{"x", "a", "b", "c"}
	if got := asStrs(t, execCmd(t, r, st, "LRANGE", "l", "0", "-1")); !reflect.DeepEqual(got, want) {
		t.Fatalf("LRANGE = %v, want %v", got, want)
	}
	wantInt(t, execCmd(t, r, st, "LLEN", "l"), 4)
}

func TestListPushX(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantInt(t, execCmd(t, r, st, "LPUSHX", "l", "a"), 0)
	wantInt(t, execCmd(t, r, st, "RPUSHX", "l", "a"), 0)
	wantInt(t, execCmd(t, r, st, "RPUSH", "l", "a"), 1)
	wantInt(t, execCmd(t, r, st, "LPUSHX", "l", "b"), 2)
}

func TestListPopIndex(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "RPUSH", "l", "a", "b", "c")
	wantBulk(t, execCmd(t, r, st, "LPOP", "l"), "a")
	wantBulk(t, execCmd(t, r, st, "RPOP", "l"), "c")
	wantBulk(t, execCmd(t, r, st, "LINDEX", "l", "0"), "b")
	wantBulk(t, execCmd(t, r, st, "LINDEX", "l", "-1"), "b")
}

func TestListPopCount(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "RPUSH", "l", "a", "b", "c")
	if got := asStrs(t, execCmd(t, r, st, "LPOP", "l", "2")); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("LPOP count = %v", got)
	}
}

func TestListSetInsertRemTrim(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "RPUSH", "l", "a", "b", "c")
	wantOK(t, execCmd(t, r, st, "LSET", "l", "1", "B"))
	wantBulk(t, execCmd(t, r, st, "LINDEX", "l", "1"), "B")
	wantInt(t, execCmd(t, r, st, "LINSERT", "l", "BEFORE", "B", "x"), 4)
	wantInt(t, execCmd(t, r, st, "LREM", "l", "1", "x"), 1)
	wantOK(t, execCmd(t, r, st, "LTRIM", "l", "0", "1"))
	wantInt(t, execCmd(t, r, st, "LLEN", "l"), 2)
}

func TestListNulls(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	if v := execCmd(t, r, st, "LPOP", "missing"); v.Type != resp.TypeBulk || !v.Null {
		t.Fatalf("LPOP missing = %c null=%v, want null bulk", v.Type, v.Null)
	}
	if v := execCmd(t, r, st, "LPOP", "missing", "2"); v.Type != resp.TypeArray || !v.Null {
		t.Fatalf("LPOP missing count = %c null=%v, want null array", v.Type, v.Null)
	}
}

func TestListCommandErrors(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SET", "k", "v")
	wantErr(t, execCmd(t, r, st, "LPUSH", "k", "a"), store.ErrWrongType.Error())
	wantErr(t, execCmd(t, r, st, "LRANGE", "l", "0"), errWrongArgs("lrange"))
}