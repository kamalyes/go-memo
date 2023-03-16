/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-16 09:22:18
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-16 09:22:18
 * @FilePath: \go-memo\resp\writer.go
 * @Description: RESP 响应序列化写入器
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

import (
	"bufio"
	"io"
	"strconv"
)

// Writer 基于带缓冲输出流的 RESP 响应序列化器
type Writer struct {
	bw *bufio.Writer
}

// NewWriter 创建响应序列化器
func NewWriter(w io.Writer) *Writer {
	return &Writer{bw: bufio.NewWriter(w)}
}

// Flush 刷出缓冲数据
func (w *Writer) Flush() error {
	return w.bw.Flush()
}

// WriteValue 按类型分发写入一个 RESP 值
func (w *Writer) WriteValue(v Value) error {
	switch v.Type {
	case TypeSimpleString:
		return w.WriteSimpleString(v.Str)
	case TypeError:
		return w.WriteError(v.Str)
	case TypeInteger:
		return w.WriteInteger(v.Int)
	case TypeBulk:
		if v.Null {
			return w.WriteNullBulk()
		}
		return w.WriteBulk(v.Str)
	case TypeArray:
		if v.Null {
			return w.WriteNullArray()
		}
		return w.WriteArray(v.Array)
	default:
		return writeRaw(w.bw, crlf)
	}
}

// WriteSimpleString 写入简单字符串
func (w *Writer) WriteSimpleString(s string) error {
	return writeRaw(w.bw, string(TypeSimpleString)+s+crlf)
}

// WriteError 写入错误信息
func (w *Writer) WriteError(msg string) error {
	return writeRaw(w.bw, string(TypeError)+msg+crlf)
}

// WriteInteger 写入整数
func (w *Writer) WriteInteger(i int64) error {
	return writeRaw(w.bw, string(TypeInteger)+strconv.FormatInt(i, 10)+crlf)
}

// WriteBulk 写入批量字符串
func (w *Writer) WriteBulk(s string) error {
	if err := writeRaw(w.bw, string(TypeBulk)+strconv.Itoa(len(s))+crlf); err != nil {
		return err
	}
	return writeRaw(w.bw, s+crlf)
}

// WriteNullBulk 写入空批量字符串
func (w *Writer) WriteNullBulk() error {
	return writeRaw(w.bw, nullBulk)
}

// WriteNullArray 写入空数组
func (w *Writer) WriteNullArray() error {
	return writeRaw(w.bw, nullArray)
}

// WriteArray 写入数组（n 元素逐一序列化）
func (w *Writer) WriteArray(v []Value) error {
	if err := writeRaw(w.bw, string(TypeArray)+strconv.Itoa(len(v))+crlf); err != nil {
		return err
	}
	for _, e := range v {
		if err := w.WriteValue(e); err != nil {
			return err
		}
	}
	return nil
}

func writeRaw(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}