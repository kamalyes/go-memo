/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 10:05:28
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 10:05:28
 * @FilePath: \go-memo\resp\marshal.go
 * @Description: RESP 命令序列化工具
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package resp

import (
	"bytes"
	"strconv"
)

// MarshalCommand 将命令参数序列化为 RESP 批量字符串数组字节，供持久化与复制流复用
func MarshalCommand(args []string) []byte {
	var b bytes.Buffer
	b.WriteByte(TypeArray)
	b.WriteString(strconv.Itoa(len(args)))
	b.WriteString(crlf)
	for _, a := range args {
		b.WriteByte(TypeBulk)
		b.WriteString(strconv.Itoa(len(a)))
		b.WriteString(crlf)
		b.WriteString(a)
		b.WriteString(crlf)
	}
	return b.Bytes()
}
