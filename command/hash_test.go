/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-16 09:12:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-16 09:12:05
 * @FilePath: \go-memo\command\hash_test.go
 * @Description: 哈希类型命令单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"reflect"
	"testing"

	"github.com/kamalyes/go-memo/store"
)

func TestHashCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantInt(t, execCmd(t, r, st, "HSET", "h", "f1", "v1"), 1)
	wantInt(t, execCmd(t, r, st, "HSET", "h", "f1", "v2"), 0)
	wantInt(t, execCmd(t, r, st, "HSET", "h", "f2", "v2"), 1)
	wantBulk(t, execCmd(t, r, st, "HGET", "h", "f1"), "v2")
	wantInt(t, execCmd(t, r, st, "HLEN", "h"), 2)
	wantInt(t, execCmd(t, r, st, "HEXISTS", "h", "f2"), 1)
	wantInt(t, execCmd(t, r, st, "HEXISTS", "h", "missing"), 0)
}

func TestHashMultiCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantOK(t, execCmd(t, r, st, "HMSET", "h", "a", "1", "b", "2"))
	wantInt(t, execCmd(t, r, st, "HSETNX", "h", "a", "x"), 0)
	wantInt(t, execCmd(t, r, st, "HSETNX", "h", "c", "3"), 1)

	if got := asStrs(t, execCmd(t, r, st, "HMGET", "h", "a", "c")); !reflect.DeepEqual(got, []string{"1", "3"}) {
		t.Fatalf("HMGET = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "HGETALL", "h")); !reflect.DeepEqual(got, []string{"a", "1", "b", "2", "c", "3"}) {
		t.Fatalf("HGETALL = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "HKEYS", "h")); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("HKEYS = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "HVALS", "h")); !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Fatalf("HVALS = %v", got)
	}
	wantInt(t, execCmd(t, r, st, "HINCRBY", "h", "n", "5"), 5)
	wantInt(t, execCmd(t, r, st, "HINCRBY", "h", "n", "-2"), 3)
	wantInt(t, execCmd(t, r, st, "HDEL", "h", "a"), 1)
}

func TestHashCommandErrors(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SET", "k", "v")
	wantErr(t, execCmd(t, r, st, "HGET", "k", "f"), store.ErrWrongType.Error())
	wantErr(t, execCmd(t, r, st, "HGET", "h"), errWrongArgs("hget"))
	wantErr(t, execCmd(t, r, st, "HINCRBY", "h", "f", "x"), msgNotInteger)
}