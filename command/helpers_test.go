/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-15 09:10:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-15 09:10:33
 * @FilePath: \go-memo\command\helpers_test.go
 * @Description: 命令层测试公共断言辅助
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"testing"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// execCmd 通过注册表分发命令，供各类型命令测试复用
func execCmd(t *testing.T, r *Registry, st store.Store, args ...string) resp.Value {
	t.Helper()
	return r.Dispatch(st, args)
}

// asStrs 将数组响应展开为批量字符串切片
func asStrs(t *testing.T, v resp.Value) []string {
	t.Helper()
	if v.Type != resp.TypeArray || v.Null {
		t.Fatalf("expected array, got type=%c null=%v", v.Type, v.Null)
	}
	out := make([]string, len(v.Array))
	for i, e := range v.Array {
		if e.Type != resp.TypeBulk || e.Null {
			t.Fatalf("expected bulk at %d, got type=%c", i, e.Type)
		}
		out[i] = e.Str
	}
	return out
}

// wantInt 断言响应为指定整数
func wantInt(t *testing.T, v resp.Value, want int64) {
	t.Helper()
	if v.Type != resp.TypeInteger || v.Int != want {
		t.Fatalf("got %c %d, want integer %d", v.Type, v.Int, want)
	}
}

// wantBulk 断言响应为指定批量字符串
func wantBulk(t *testing.T, v resp.Value, want string) {
	t.Helper()
	if v.Type != resp.TypeBulk || v.Null || v.Str != want {
		t.Fatalf("got %c %q null=%v, want bulk %q", v.Type, v.Str, v.Null, want)
	}
}

// wantOK 断言响应为 OK
func wantOK(t *testing.T, v resp.Value) {
	t.Helper()
	if v.Type != resp.TypeSimpleString || v.Str != msgOK {
		t.Fatalf("got %c %q, want OK", v.Type, v.Str)
	}
}

// wantErr 断言响应为指定错误文本
func wantErr(t *testing.T, v resp.Value, want string) {
	t.Helper()
	if v.Type != resp.TypeError || v.Str != want {
		t.Fatalf("got %c %q, want error %q", v.Type, v.Str, want)
	}
}