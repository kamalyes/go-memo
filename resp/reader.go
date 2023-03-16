/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-16 09:12:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-16 09:12:33
 * @FilePath: \go-memo\resp\reader.go
 * @Description: RESP 流式协议解析读取器
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// errInvalidLine 行终止符缺失或格式非法
var errInvalidLine = errors.New("resp: invalid line terminator")

// ReadValue 从流中读取一个 RESP 值
func ReadValue(r *bufio.Reader) (Value, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch prefix {
	case TypeSimpleString:
		return readSimpleString(r)
	case TypeError:
		return readSimpleStringAs(r, TypeError)
	case TypeInteger:
		return readInteger(r)
	case TypeBulk:
		return readBulk(r)
	case TypeArray:
		return readArray(r)
	default:
		return Value{}, fmt.Errorf("resp: unknown type byte %q", prefix)
	}
}

// ReadCommand 读取一条客户端命令，命令必须是批量字符串数组
func ReadCommand(r *bufio.Reader) ([]string, error) {
	v, err := ReadValue(r)
	if err != nil {
		return nil, err
	}
	if v.Type != TypeArray {
		return nil, fmt.Errorf("resp: expected array, got type %q", v.Type)
	}
	args := make([]string, 0, len(v.Array))
	for _, e := range v.Array {
		if e.Type != TypeBulk {
			return nil, fmt.Errorf("resp: expected bulk string element, got type %q", e.Type)
		}
		args = append(args, e.Str)
	}
	return args, nil
}

// readLine 读取以 CRLF 结尾的一行，返回去除终止符的内容
func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", errInvalidLine
	}
	return string(line[:len(line)-2]), nil
}

func readSimpleString(r *bufio.Reader) (Value, error) {
	s, err := readLine(r)
	if err != nil {
		return Value{}, err
	}
	return SimpleString(s), nil
}

func readSimpleStringAs(r *bufio.Reader, typ byte) (Value, error) {
	s, err := readLine(r)
	if err != nil {
		return Value{}, err
	}
	return Value{Type: typ, Str: s}, nil
}

func readInteger(r *bufio.Reader) (Value, error) {
	line, err := readLine(r)
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, fmt.Errorf("resp: invalid integer %q", line)
	}
	return Integer(n), nil
}

func readBulk(r *bufio.Reader) (Value, error) {
	line, err := readLine(r)
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, fmt.Errorf("resp: invalid bulk length %q", line)
	}
	if n == -1 {
		return NullBulk(), nil
	}
	if n < 0 {
		return Value{}, fmt.Errorf("resp: invalid bulk length %d", n)
	}
	buf := make([]byte, n+2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Value{}, err
	}
	if buf[n] != '\r' || buf[n+1] != '\n' {
		return Value{}, errInvalidLine
	}
	return BulkString(string(buf[:n])), nil
}

func readArray(r *bufio.Reader) (Value, error) {
	line, err := readLine(r)
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, fmt.Errorf("resp: invalid array length %q", line)
	}
	if n == -1 {
		return NullArray(), nil
	}
	if n < 0 {
		return Value{}, fmt.Errorf("resp: invalid array length %d", n)
	}
	elems := make([]Value, 0, n)
	for i := int64(0); i < n; i++ {
		e, err := ReadValue(r)
		if err != nil {
			return Value{}, err
		}
		elems = append(elems, e)
	}
	return Array(elems...), nil
}