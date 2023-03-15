/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-15 10:25:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-15 10:25:33
 * @FilePath: \go-memo\resp\value.go
 * @Description: RESP 值类型与构造器定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

// Value 一个已解析的 RESP 值，通过 Type 区分具体形态
type Value struct {
	Type  byte    // 类型首字节
	Str   string  // 简单字符串 / 错误文本 / 批量字符串内容
	Int   int64   // 整数
	Array []Value // 数组元素
	Null  bool    // 空批量字符串或空数组标记
}

// SimpleString 构造简单字符串值
func SimpleString(s string) Value {
	return Value{Type: TypeSimpleString, Str: s}
}

// ErrorString 构造错误值
func ErrorString(s string) Value {
	return Value{Type: TypeError, Str: s}
}

// Integer 构造整数值
func Integer(i int64) Value {
	return Value{Type: TypeInteger, Int: i}
}

// BulkString 构造批量字符串值
func BulkString(s string) Value {
	return Value{Type: TypeBulk, Str: s}
}

// NullBulk 构造空批量字符串
func NullBulk() Value {
	return Value{Type: TypeBulk, Null: true}
}

// Array 构造数组值
func Array(v ...Value) Value {
	return Value{Type: TypeArray, Array: v}
}

// NullArray 构造空数组
func NullArray() Value {
	return Value{Type: TypeArray, Null: true}
}