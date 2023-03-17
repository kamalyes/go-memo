/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-16 11:38:22
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-16 11:38:22
 * @FilePath: \go-memo\resp\writer_test.go
 * @Description: RESP 序列化写入器单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestWriteSimpleString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteSimpleString("OK"); err != nil {
		t.Fatalf("write simple string: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := buf.String(); got != "+OK\r\n" {
		t.Fatalf("got %q, want %q", got, "+OK\r\n")
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteError("ERR unknown command"); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := buf.String(); got != "-ERR unknown command\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteInteger(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteInteger(1000); err != nil {
		t.Fatalf("write integer: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := buf.String(); got != ":1000\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteBulk(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteBulk("hello"); err != nil {
		t.Fatalf("write bulk: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := buf.String(); got != "$5\r\nhello\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteNullBulk(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteValue(NullBulk()); err != nil {
		t.Fatalf("write null bulk: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := buf.String(); got != "$-1\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteArray(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteArray([]Value{BulkString("a"), Integer(1)}); err != nil {
		t.Fatalf("write array: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	want := "*2\r\n$1\r\na\r\n:1\r\n"
	if got := buf.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	payload := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	args, err := ReadCommand(bufio.NewReader(strings.NewReader(payload)))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(args) != 3 || args[0] != "SET" {
		t.Fatalf("unexpected args: %v", args)
	}
}