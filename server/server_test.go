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
	"io"
	"net"
	"testing"
	"time"
)

func startServer(t *testing.T) (string, *Server) {
	t.Helper()
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