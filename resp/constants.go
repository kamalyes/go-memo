/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-15 10:22:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-15 10:22:11
 * @FilePath: \go-memo\resp\constants.go
 * @Description: RESP 协议类型首字节与分隔符常量定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

// 类型首字节：每个 RESP 值以其类型标识作为首个字节
const (
	TypeSimpleString = '+' // 简单字符串
	TypeError        = '-' // 错误信息
	TypeInteger      = ':' // 整数
	TypeBulk         = '$' // 批量字符串
	TypeArray        = '*' // 数组
)

// protocol 级分隔符与空值标记
const (
	crlf      = "\r\n"   // RESP 行终止符
	nullBulk  = "$-1\r\n" // 空批量字符串
	nullArray = "*-1\r\n" // 空数组
)