/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-16 11:32:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-16 11:32:05
 * @FilePath: \go-memo\resp\reader_test.go
 * @Description: RESP 解析读取器单元测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadSimpleString(t *testing.T) {
	v, err := ReadValue(bufio.NewReader(strings.NewReader("+OK\r\n")))
	if err != nil {
		t.Fatalf("read simple string: %v", err)
	}
	if v.Type != TypeSimpleString || v.Str != "OK" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestReadError(t *testing.T) {
	v, err := ReadValue(bufio.NewReader(strings.NewReader("-ERR unknown command\r\n")))
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if v.Type != TypeError || v.Str != "ERR unknown command" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestReadInteger(t *testing.T) {
	v, err := ReadValue(bufio.NewReader(strings.NewReader(":1000\r\n")))
	if err != nil {
		t.Fatalf("read integer: %v", err)
	}
	if v.Type != TypeInteger || v.Int != 1000 {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestReadBulk(t *testing.T) {
	v, err := ReadValue(bufio.NewReader(strings.NewReader("$5\r\nhello\r\n")))
	if err != nil {
		t.Fatalf("read bulk: %v", err)
	}
	if v.Type != TypeBulk || v.Str != "hello" || v.Null {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestReadNullBulk(t *testing.T) {
	v, err := ReadValue(bufio.NewReader(strings.NewReader("$-1\r\n")))
	if err != nil {
		t.Fatalf("read null bulk: %v", err)
	}
	if v.Type != TypeBulk || !v.Null {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestReadCommand(t *testing.T) {
	args, err := ReadCommand(bufio.NewReader(strings.NewReader("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")))
	if err != nil {
		t.Fatalf("read command: %v", err)
	}
	want := []string{"SET", "foo", "bar"}
	if len(args) != len(want) {
		t.Fatalf("arg count = %d, want %d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestReadInvalidLine(t *testing.T) {
	if _, err := ReadValue(bufio.NewReader(strings.NewReader("+OK\n"))); err == nil {
		t.Fatal("expected error for missing CR")
	}
}