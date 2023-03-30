/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 11:33:07
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-19 11:33:07
 * @FilePath: \go-memo\server\server_test.go
 * @Description: 服务 TCP 闭环与优雅关闭测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"bufio"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamalyes/go-memo/resp"
)

func startServer(t *testing.T) (string, *Server) {
	return startServerWith(t, WithAddr("127.0.0.1:0"))
}

func startServerWith(t *testing.T, opts ...Option) (string, *Server) {
	t.Helper()
	s := New(opts...)
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	for i := 0; i < 200; i++ {
		if s.Addr() != nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if s.Addr() == nil {
		t.Fatal("server did not start")
	}
	return s.Addr().String(), s
}

func roundTrip(t *testing.T, addr, req, want string) {
	t.Helper()
	nc, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer nc.Close()
	if _, err := nc.Write([]byte(req)); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, len(want))
	if _, err := io.ReadFull(nc, buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := string(buf); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPing(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()
	roundTrip(t, addr, "*1\r\n$4\r\nPING\r\n", "+PONG\r\n")
}

func TestSetGetAcrossConnections(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()
	roundTrip(t, addr, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", "+OK\r\n")
	roundTrip(t, addr, "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n", "$3\r\nbar\r\n")
}

func TestGetMissing(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()
	roundTrip(t, addr, "*2\r\n$3\r\nGET\r\n$7\r\nmissing\r\n", "$-1\r\n")
}

func TestGracefulClose(t *testing.T) {
	s := New(WithAddr("127.0.0.1:0"))
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	for i := 0; i < 200; i++ {
		if s.Addr() != nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if s.Addr() == nil {
		t.Fatal("server did not start")
	}
	roundTrip(t, s.Addr().String(), "*1\r\n$4\r\nPING\r\n", "+PONG\r\n")
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("ListenAndServe = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

// client 测试用持久连接，可串行收发多条命令
type client struct {
	t  testing.TB
	nc net.Conn
	br *bufio.Reader
}

func dialClient(t testing.TB, addr string) *client {
	t.Helper()
	nc, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return &client{t: t, nc: nc, br: bufio.NewReader(nc)}
}

func (c *client) close() { _ = c.nc.Close() }

// cmd 发送一条命令并读取一条响应
func (c *client) cmd(args ...string) resp.Value {
	if _, err := c.nc.Write(resp.MarshalCommand(args)); err != nil {
		c.t.Fatalf("write: %v", err)
	}
	v, err := resp.ReadValue(c.br)
	if err != nil {
		c.t.Fatalf("read: %v", err)
	}
	return v
}

func TestSelectMultiDB(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()
	c := dialClient(t, addr)
	defer c.close()

	if v := c.cmd("SELECT", "1"); v.Type != resp.TypeSimpleString || v.Str != "OK" {
		t.Fatalf("SELECT 1 = %+v, want OK", v)
	}
	if v := c.cmd("SET", "k", "v1"); v.Str != "OK" {
		t.Fatalf("SET in db1 = %+v", v)
	}
	if v := c.cmd("SELECT", "0"); v.Str != "OK" {
		t.Fatalf("SELECT 0 = %+v", v)
	}
	if v := c.cmd("GET", "k"); v.Type != resp.TypeBulk || !v.Null {
		t.Fatalf("GET k in db0 = %+v, want null", v)
	}
	if v := c.cmd("SELECT", "1"); v.Str != "OK" {
		t.Fatalf("SELECT back to 1 = %+v", v)
	}
	if v := c.cmd("GET", "k"); v.Type != resp.TypeBulk || v.Str != "v1" {
		t.Fatalf("GET k in db1 = %+v, want v1", v)
	}
	if v := c.cmd("SELECT", "15"); v.Str != "OK" {
		t.Fatalf("SELECT 15 = %+v", v)
	}
	if v := c.cmd("SELECT", "16"); v.Type != resp.TypeError {
		t.Fatalf("SELECT 16 = %+v, want error", v)
	}
}

func TestClientListAndInfo(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()

	c1 := dialClient(t, addr)
	defer c1.close()
	c2 := dialClient(t, addr)
	defer c2.close()

	c1.cmd("CLIENT", "SETNAME", "alice")
	if v := c1.cmd("CLIENT", "GETNAME"); v.Type != resp.TypeBulk || v.Str != "alice" {
		t.Fatalf("GETNAME = %+v, want alice", v)
	}

	v := c1.cmd("CLIENT", "LIST")
	if v.Type != resp.TypeBulk {
		t.Fatalf("CLIENT LIST = %+v, want bulk", v)
	}
	if !strings.Contains(v.Str, "addr=") || !strings.Contains(v.Str, "name=alice") {
		t.Fatalf("CLIENT LIST missing fields: %q", v.Str)
	}
	if strings.Count(v.Str, "id=") < 2 {
		t.Fatalf("CLIENT LIST should contain at least 2 clients: %q", v.Str)
	}

	info := c1.cmd("INFO")
	if info.Type != resp.TypeBulk {
		t.Fatalf("INFO = %+v, want bulk", info)
	}
	if !strings.Contains(info.Str, "redis_version:") {
		t.Fatalf("INFO missing redis_version: %q", info.Str)
	}
	if !strings.Contains(info.Str, "uptime_in_seconds:") {
		t.Fatalf("INFO missing uptime_in_seconds: %q", info.Str)
	}
	if !strings.Contains(info.Str, "databases:16") {
		t.Fatalf("INFO missing databases:16: %q", info.Str)
	}
	if !strings.Contains(info.Str, "connected_clients:") {
		t.Fatalf("INFO missing connected_clients: %q", info.Str)
	}
}

func TestFlushDBAndFlushAll(t *testing.T) {
	addr, s := startServer(t)
	defer s.Close()
	c := dialClient(t, addr)
	defer c.close()

	c.cmd("SET", "a", "1")
	c.cmd("SELECT", "1")
	c.cmd("SET", "b", "2")

	if v := c.cmd("DBSIZE"); v.Int != 1 {
		t.Fatalf("DBSIZE db1 = %+v, want 1", v)
	}
	if v := c.cmd("FLUSHDB"); v.Str != "OK" {
		t.Fatalf("FLUSHDB = %+v", v)
	}
	if v := c.cmd("DBSIZE"); v.Int != 0 {
		t.Fatalf("DBSIZE after FLUSHDB = %+v, want 0", v)
	}

	if v := c.cmd("SELECT", "0"); v.Str != "OK" {
		t.Fatalf("SELECT 0 = %+v", v)
	}
	if v := c.cmd("DBSIZE"); v.Int != 1 {
		t.Fatalf("DBSIZE db0 = %+v, want 1", v)
	}

	if v := c.cmd("FLUSHALL"); v.Str != "OK" {
		t.Fatalf("FLUSHALL = %+v", v)
	}
	if v := c.cmd("DBSIZE"); v.Int != 0 {
		t.Fatalf("DBSIZE after FLUSHALL = %+v, want 0", v)
	}
}

func TestAOFPersistRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	addr, s := startServerWith(t, WithAddr("127.0.0.1:0"), WithAOF(path))
	c := dialClient(t, addr)
	if v := c.cmd("SET", "foo", "bar"); v.Str != "OK" {
		t.Fatalf("SET = %+v", v)
	}
	c.close()
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	addr, s = startServerWith(t, WithAddr("127.0.0.1:0"), WithAOF(path))
	defer s.Close()
	c = dialClient(t, addr)
	defer c.close()
	if v := c.cmd("GET", "foo"); v.Type != resp.TypeBulk || v.Str != "bar" {
		t.Fatalf("GET after restart = %+v, want bar", v)
	}
}
