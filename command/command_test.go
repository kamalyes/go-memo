/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 11:26:30
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 11:26:30
 * @FilePath: \go-memo\command\command_test.go
 * @Description: 命令分发与语义单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"testing"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// env 命令测试环境，共享同一注册表与存储后端
type env struct {
	reg *Registry
	st  store.Store
}

func newEnv() *env {
	return &env{reg: NewRegistry(), st: store.New()}
}

func (e *env) cmd(args ...string) resp.Value {
	return e.reg.Dispatch(e.st, args)
}

func TestPing(t *testing.T) {
	v := newEnv().cmd("PING")
	if v.Type != resp.TypeSimpleString || v.Str != "PONG" {
		t.Fatalf("PING = %+v, want PONG", v)
	}
}

func TestSetGetRoundTrip(t *testing.T) {
	e := newEnv()
	if v := e.cmd("SET", "foo", "bar"); v.Type != resp.TypeSimpleString || v.Str != "OK" {
		t.Fatalf("SET = %+v", v)
	}
	v := e.cmd("GET", "foo")
	if v.Type != resp.TypeBulk || v.Str != "bar" {
		t.Fatalf("GET = %+v, want bar", v)
	}
}

func TestGetMissing(t *testing.T) {
	v := newEnv().cmd("GET", "missing")
	if v.Type != resp.TypeBulk || !v.Null {
		t.Fatalf("GET missing = %+v, want null bulk", v)
	}
}

func TestIncrBy(t *testing.T) {
	e := newEnv()
	if v := e.cmd("INCRBY", "cnt", "10"); v.Type != resp.TypeInteger || v.Int != 10 {
		t.Fatalf("INCRBY = %+v, want 10", v)
	}
	if v := e.cmd("INCRBY", "cnt", "-3"); v.Int != 7 {
		t.Fatalf("INCRBY = %+v, want 7", v)
	}
}

func TestIncrByNotInteger(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "str", "abc")
	if v := e.cmd("INCRBY", "str", "1"); v.Type != resp.TypeError {
		t.Fatalf("INCRBY non-int = %+v, want error", v)
	}
}

func TestDel(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "a", "1")
	e.cmd("SET", "b", "2")
	if v := e.cmd("DEL", "a", "b", "missing"); v.Type != resp.TypeInteger || v.Int != 2 {
		t.Fatalf("DEL = %+v, want 2", v)
	}
}

func TestExpireTTL(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "k", "v")
	if v := e.cmd("EXPIRE", "k", "60"); v.Int != 1 {
		t.Fatalf("EXPIRE = %+v, want 1", v)
	}
	if v := e.cmd("TTL", "k"); v.Int != 60 {
		t.Fatalf("TTL = %+v, want 60", v)
	}
}

func TestPExpirePTTL(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "k", "v")
	if v := e.cmd("PEXPIRE", "k", "60000"); v.Int != 1 {
		t.Fatalf("PEXPIRE = %+v, want 1", v)
	}
	v := e.cmd("PTTL", "k")
	if v.Int <= 0 || v.Int > 60000 {
		t.Fatalf("PTTL = %+v, want in (0, 60000]", v)
	}
}

func TestTTLMissingAndNoExpire(t *testing.T) {
	e := newEnv()
	if v := e.cmd("TTL", "missing"); v.Int != -2 {
		t.Fatalf("TTL missing = %+v, want -2", v)
	}
	e.cmd("SET", "k", "v")
	if v := e.cmd("TTL", "k"); v.Int != -1 {
		t.Fatalf("TTL no-expire = %+v, want -1", v)
	}
}

func TestUnknownCommand(t *testing.T) {
	if v := newEnv().cmd("NOPE"); v.Type != resp.TypeError {
		t.Fatalf("unknown command = %+v, want error", v)
	}
}

func TestCaseInsensitive(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "foo", "bar")
	if v := e.cmd("Get", "foo"); v.Type != resp.TypeBulk || v.Str != "bar" {
		t.Fatalf("case-insensitive GET = %+v", v)
	}
}

func TestDBSize(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "a", "1")
	e.cmd("SET", "b", "2")
	if v := e.cmd("DBSIZE"); v.Type != resp.TypeInteger || v.Int != 2 {
		t.Fatalf("DBSIZE = %+v, want 2", v)
	}
}

func TestType(t *testing.T) {
	e := newEnv()
	if v := e.cmd("TYPE", "missing"); v.Type != resp.TypeSimpleString || v.Str != "none" {
		t.Fatalf("TYPE missing = %+v, want none", v)
	}
	e.cmd("SET", "k", "v")
	if v := e.cmd("TYPE", "k"); v.Type != resp.TypeSimpleString || v.Str != "string" {
		t.Fatalf("TYPE k = %+v, want string", v)
	}
}

func TestScan(t *testing.T) {
	e := newEnv()
	for _, k := range []string{"a", "b", "c"} {
		e.cmd("SET", k, "v")
	}
	v := e.cmd("SCAN", "0")
	if v.Type != resp.TypeArray || len(v.Array) != 2 {
		t.Fatalf("SCAN = %+v, want 2-element array", v)
	}
	keys := v.Array[1]
	if keys.Type != resp.TypeArray || len(keys.Array) != 3 {
		t.Fatalf("SCAN keys = %+v, want 3 keys", keys)
	}
}

func TestScanMatch(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "user:1", "a")
	e.cmd("SET", "user:2", "b")
	e.cmd("SET", "order:1", "c")
	v := e.cmd("SCAN", "0", "MATCH", "user:*")
	if v.Type != resp.TypeArray || len(v.Array) != 2 {
		t.Fatalf("SCAN MATCH = %+v, want 2-element array", v)
	}
	if keys := v.Array[1]; len(keys.Array) != 2 {
		t.Fatalf("SCAN MATCH keys = %+v, want 2 keys", keys)
	}
}

func TestKeys(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "foo", "1")
	e.cmd("SET", "bar", "2")
	v := e.cmd("KEYS", "*")
	if v.Type != resp.TypeArray || len(v.Array) != 2 {
		t.Fatalf("KEYS * = %+v, want 2 keys", v)
	}
}

func TestSetOptions(t *testing.T) {
	e := newEnv()
	if v := e.cmd("SET", "nx", "1", "NX"); v.Type != resp.TypeSimpleString || v.Str != "OK" {
		t.Fatalf("SET nx = %+v", v)
	}
	if v := e.cmd("SET", "nx", "2", "NX"); v.Type != resp.TypeBulk || !v.Null {
		t.Fatalf("SET nx NX again = %+v, want null", v)
	}
	if v := e.cmd("SET", "nx", "3", "XX"); v.Type != resp.TypeSimpleString {
		t.Fatalf("SET nx XX = %+v", v)
	}
	if v := e.cmd("GET", "nx"); v.Str != "3" {
		t.Fatalf("GET nx = %+v, want 3", v)
	}
	if v := e.cmd("SET", "gone", "1", "XX"); v.Type != resp.TypeBulk || !v.Null {
		t.Fatalf("SET xx missing = %+v, want null", v)
	}
	if v := e.cmd("SET", "ex", "v", "EX", "100"); v.Str != "OK" {
		t.Fatalf("SET ex = %+v", v)
	}
	if v := e.cmd("TTL", "ex"); v.Int != 100 {
		t.Fatalf("TTL ex = %+v, want 100", v)
	}
}

func TestIncrDecr(t *testing.T) {
	e := newEnv()
	if v := e.cmd("INCR", "n"); v.Int != 1 {
		t.Fatalf("INCR = %+v", v)
	}
	if v := e.cmd("INCR", "n"); v.Int != 2 {
		t.Fatalf("INCR = %+v", v)
	}
	if v := e.cmd("DECR", "n"); v.Int != 1 {
		t.Fatalf("DECR = %+v", v)
	}
	if v := e.cmd("DECRBY", "n", "5"); v.Int != -4 {
		t.Fatalf("DECRBY = %+v", v)
	}
}

func TestAppendStrLen(t *testing.T) {
	e := newEnv()
	if v := e.cmd("APPEND", "s", "hello"); v.Int != 5 {
		t.Fatalf("APPEND = %+v", v)
	}
	if v := e.cmd("APPEND", "s", " world"); v.Int != 11 {
		t.Fatalf("APPEND = %+v", v)
	}
	if v := e.cmd("STRLEN", "s"); v.Int != 11 {
		t.Fatalf("STRLEN = %+v", v)
	}
	if v := e.cmd("STRLEN", "missing"); v.Int != 0 {
		t.Fatalf("STRLEN missing = %+v", v)
	}
}

func TestMGetMSet(t *testing.T) {
	e := newEnv()
	if v := e.cmd("MSET", "a", "1", "b", "2"); v.Str != "OK" {
		t.Fatalf("MSET = %+v", v)
	}
	v := e.cmd("MGET", "a", "missing", "b")
	if v.Type != resp.TypeArray || len(v.Array) != 3 {
		t.Fatalf("MGET = %+v", v)
	}
	if v.Array[0].Str != "1" || !v.Array[1].Null || v.Array[2].Str != "2" {
		t.Fatalf("MGET values = %+v", v)
	}
}

func TestSetExSetNxPersist(t *testing.T) {
	e := newEnv()
	if v := e.cmd("SETEX", "k", "100", "v"); v.Str != "OK" {
		t.Fatalf("SETEX = %+v", v)
	}
	if v := e.cmd("TTL", "k"); v.Int != 100 {
		t.Fatalf("TTL after SETEX = %+v", v)
	}
	if v := e.cmd("PERSIST", "k"); v.Int != 1 {
		t.Fatalf("PERSIST = %+v", v)
	}
	if v := e.cmd("TTL", "k"); v.Int != -1 {
		t.Fatalf("TTL after PERSIST = %+v, want -1", v)
	}
	if v := e.cmd("SETNX", "k", "x"); v.Int != 0 {
		t.Fatalf("SETNX existing = %+v, want 0", v)
	}
	if v := e.cmd("SETNX", "new", "x"); v.Int != 1 {
		t.Fatalf("SETNX new = %+v, want 1", v)
	}
}

func TestExists(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "a", "1")
	e.cmd("SET", "b", "2")
	if v := e.cmd("EXISTS", "a", "missing", "b"); v.Int != 2 {
		t.Fatalf("EXISTS = %+v, want 2", v)
	}
}

func TestRename(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "src", "val")
	if v := e.cmd("RENAME", "src", "dst"); v.Str != "OK" {
		t.Fatalf("RENAME = %+v", v)
	}
	if v := e.cmd("GET", "src"); !v.Null {
		t.Fatalf("GET src = %+v, want null", v)
	}
	if v := e.cmd("GET", "dst"); v.Str != "val" {
		t.Fatalf("GET dst = %+v, want val", v)
	}
	if v := e.cmd("RENAME", "missing", "x"); v.Type != resp.TypeError {
		t.Fatalf("RENAME missing = %+v, want error", v)
	}
}

func TestRenamePreservesTTL(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "src", "val", "EX", "100")
	if v := e.cmd("RENAME", "src", "dst"); v.Str != "OK" {
		t.Fatalf("RENAME = %+v", v)
	}
	if v := e.cmd("TTL", "dst"); v.Int != 100 {
		t.Fatalf("TTL dst = %+v, want 100", v)
	}
}

func TestDumpRestore(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "src", "hello")
	v := e.cmd("DUMP", "src")
	if v.Type != resp.TypeBulk {
		t.Fatalf("DUMP = %+v, want bulk", v)
	}
	if r := e.cmd("RESTORE", "dst", "0", v.Str, "REPLACE"); r.Str != "OK" {
		t.Fatalf("RESTORE = %+v", r)
	}
	if r := e.cmd("GET", "dst"); r.Str != "hello" {
		t.Fatalf("GET dst = %+v, want hello", r)
	}
	if r := e.cmd("DUMP", "missing"); r.Type != resp.TypeBulk || !r.Null {
		t.Fatalf("DUMP missing = %+v, want null", r)
	}
}

func TestRestoreTtlAndBusyKey(t *testing.T) {
	e := newEnv()
	e.cmd("SET", "k", "v")
	dump := e.cmd("DUMP", "k")
	if r := e.cmd("RESTORE", "r", "100000", dump.Str); r.Str != "OK" {
		t.Fatalf("RESTORE = %+v", r)
	}
	if r := e.cmd("TTL", "r"); r.Int != 100 {
		t.Fatalf("TTL r = %+v, want 100", r)
	}
	if r := e.cmd("RESTORE", "r", "0", dump.Str); r.Type != resp.TypeError {
		t.Fatalf("RESTORE busy = %+v, want error", r)
	}
}
