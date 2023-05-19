/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-16 09:21:52
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-16 09:21:52
 * @FilePath: \go-memo\command\zset_test.go
 * @Description: 有序集合类型命令单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"reflect"
	"testing"

	"github.com/kamalyes/go-memo/store"
)

func TestZSetCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantInt(t, execCmd(t, r, st, "ZADD", "z", "1", "a", "2", "b", "1.5", "c"), 3)
	wantInt(t, execCmd(t, r, st, "ZADD", "z", "5", "a"), 0)
	wantInt(t, execCmd(t, r, st, "ZCARD", "z"), 3)
	wantBulk(t, execCmd(t, r, st, "ZSCORE", "z", "a"), "5")

	if got := asStrs(t, execCmd(t, r, st, "ZRANGE", "z", "0", "-1")); !reflect.DeepEqual(got, []string{"c", "b", "a"}) {
		t.Fatalf("ZRANGE = %v", got)
	}
	if got := asStrs(t, execCmd(t, r, st, "ZREVRANGE", "z", "0", "-1")); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("ZREVRANGE = %v", got)
	}
}

func TestZSetScoreRangeCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	wantBulk(t, execCmd(t, r, st, "ZINCRBY", "z", "1", "a"), "1")
	wantBulk(t, execCmd(t, r, st, "ZINCRBY", "z", "1", "a"), "2")
	execCmd(t, r, st, "ZADD", "z", "2", "b", "3", "c")

	if got := asStrs(t, execCmd(t, r, st, "ZRANGEBYSCORE", "z", "(2", "+inf")); !reflect.DeepEqual(got, []string{"c"}) {
		t.Fatalf("ZRANGEBYSCORE = %v", got)
	}
	wantInt(t, execCmd(t, r, st, "ZCOUNT", "z", "1", "2"), 2)
	wantInt(t, execCmd(t, r, st, "ZRANK", "z", "a"), 0)
	wantInt(t, execCmd(t, r, st, "ZREVRANK", "z", "c"), 0)
}

func TestZSetRemRangeCommands(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "ZADD", "z", "1", "a", "2", "b", "3", "c")
	wantInt(t, execCmd(t, r, st, "ZREMRANGEBYSCORE", "z", "2", "+inf"), 2)
	wantInt(t, execCmd(t, r, st, "ZCARD", "z"), 1)
	wantInt(t, execCmd(t, r, st, "ZREM", "z", "a"), 1)
}

func TestZSetCommandErrors(t *testing.T) {
	r := NewRegistry()
	st := store.New()
	execCmd(t, r, st, "SET", "k", "v")
	wantErr(t, execCmd(t, r, st, "ZADD", "k", "1", "a"), store.ErrWrongType.Error())
	wantErr(t, execCmd(t, r, st, "ZADD", "z", "abc", "a"), msgNotFloat)
	wantErr(t, execCmd(t, r, st, "ZRANGE", "z", "0"), errWrongArgs("zrange"))
}