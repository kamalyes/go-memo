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
	"strings"
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

func TestSelect(t *testing.T) {
	e := newEnv()
	if v := e.cmd("SELECT", "0"); v.Type != resp.TypeSimpleString || v.Str != "OK" {
		t.Fatalf("SELECT 0 = %+v, want OK", v)
	}
	if v := e.cmd("SELECT", "1"); v.Type != resp.TypeError {
		t.Fatalf("SELECT 1 = %+v, want error", v)
	}
}

func TestInfo(t *testing.T) {
	v := newEnv().cmd("INFO")
	if v.Type != resp.TypeBulk {
		t.Fatalf("INFO = %+v, want bulk", v)
	}
	if !strings.Contains(v.Str, "memo_version:"+Version) {
		t.Fatalf("INFO missing memo_version: %q", v.Str)
	}
	if !strings.Contains(v.Str, "redis_version:") {
		t.Fatalf("INFO missing redis_version: %q", v.Str)
	}
	if !strings.Contains(v.Str, "# Clients") {
		t.Fatalf("INFO missing Clients section: %q", v.Str)
	}
	if !strings.Contains(v.Str, "connected_clients:") {
		t.Fatalf("INFO missing connected_clients: %q", v.Str)
	}
	if !strings.Contains(v.Str, "# Memory") {
		t.Fatalf("INFO missing Memory section: %q", v.Str)
	}
	if !strings.Contains(v.Str, "used_memory:") {
		t.Fatalf("INFO missing used_memory: %q", v.Str)
	}
	if !strings.Contains(v.Str, "tcp_port:") {
		t.Fatalf("INFO missing tcp_port: %q", v.Str)
	}
	if !strings.Contains(v.Str, "# Keyspace") {
		t.Fatalf("INFO missing Keyspace section: %q", v.Str)
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
