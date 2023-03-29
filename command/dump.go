/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-29 15:26:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-29 15:26:00
 * @FilePath: \go-memo\command\dump.go
 * @Description: DUMP / RESTORE 序列化命令，供桌面客户端跨库迁移键值
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// dumpTypeString DUMP 载荷中的字符串类型标识
const dumpTypeString = 0x00

// handleDump 序列化键值为不透明字节流，键不存在返回空批量字符串
func handleDump(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("dump"))
	}
	v, ok := st.Get(args[1])
	if !ok {
		return resp.NullBulk()
	}
	return resp.BulkString(dumpPayload(v))
}

// handleRestore 反序列化 DUMP 字节流并写入键，支持 REPLACE / ABSTTL 选项
func handleRestore(st store.Store, args []string) resp.Value {
	if len(args) < 4 {
		return resp.ErrorString(errWrongArgs("restore"))
	}
	key := args[1]
	ttl, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if ttl < 0 {
		return resp.ErrorString("ERR Invalid TTL value, must be >= 0")
	}
	payload := args[3]

	replace, absttl := false, false
	for i := 4; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "REPLACE":
			replace = true
		case "ABSTTL":
			absttl = true
		case "IDLETIME", "FREQ":
			i++
			if i >= len(args) {
				return resp.ErrorString(msgSyntaxError)
			}
		default:
			return resp.ErrorString(msgSyntaxError)
		}
	}

	if !replace && st.Exists(key) {
		return resp.ErrorString("BUSYKEY Target key name already exists")
	}
	value, ok := parseDumpPayload(payload)
	if !ok {
		return resp.ErrorString("ERR DUMP payload version or checksum are wrong")
	}
	st.Set(key, value)
	if ttl > 0 {
		if absttl {
			st.Expire(key, ttl)
		} else {
			st.Expire(key, time.Now().UnixMilli()+ttl)
		}
	}
	return resp.SimpleString(msgOK)
}

// dumpPayload 构造 DUMP 载荷，格式为 [类型][长度前缀值][0xFF][2 字节 CRC16]
func dumpPayload(value string) string {
	b := make([]byte, 0, len(value)+5)
	b = append(b, dumpTypeString)
	b = appendRdbString(b, []byte(value))
	b = append(b, 0xFF)
	crc := crc16(b)
	b = append(b, byte(crc>>8), byte(crc))
	return string(b)
}

// parseDumpPayload 解析 DUMP 载荷并校验 CRC，返回原始字节值
func parseDumpPayload(payload string) (string, bool) {
	b := []byte(payload)
	if len(b) < 5 || b[0] != dumpTypeString {
		return "", false
	}
	if crc16(b[:len(b)-2]) != uint16(b[len(b)-2])<<8|uint16(b[len(b)-1]) {
		return "", false
	}
	if b[len(b)-3] != 0xFF {
		return "", false
	}
	body := b[1 : len(b)-3]
	val, n, ok := decodeRdbString(body)
	if !ok || n != len(body) {
		return "", false
	}
	return string(val), true
}

// appendRdbString 以 RDB 长度前缀编码追加字符串
func appendRdbString(dst, s []byte) []byte {
	switch n := len(s); {
	case n < 1<<6:
		dst = append(dst, byte(n))
	case n < 1<<14:
		dst = append(dst, byte(n>>8)|0x40, byte(n))
	default:
		dst = append(dst, 0x80, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	return append(dst, s...)
}

// decodeRdbString 解码 RDB 长度前缀字符串，返回字节值与消费长度
func decodeRdbString(b []byte) ([]byte, int, bool) {
	if len(b) == 0 {
		return nil, 0, false
	}
	first := b[0]
	switch {
	case first&0xC0 == 0x00:
		n := int(first)
		if len(b) < 1+n {
			return nil, 0, false
		}
		return b[1 : 1+n], 1 + n, true
	case first&0xC0 == 0x40:
		if len(b) < 2 {
			return nil, 0, false
		}
		n := int(first&0x3F)<<8 | int(b[1])
		if len(b) < 2+n {
			return nil, 0, false
		}
		return b[2 : 2+n], 2 + n, true
	case first&0xC0 == 0x80:
		if len(b) < 5 {
			return nil, 0, false
		}
		n := int(b[1])<<24 | int(b[2])<<16 | int(b[3])<<8 | int(b[4])
		if len(b) < 5+n {
			return nil, 0, false
		}
		return b[5 : 5+n], 5 + n, true
	default:
		return nil, 0, false
	}
}

// crc16 计算 CCITT (XMODEM) 十六位校验和，与 Redis DUMP 载荷对齐
func crc16(b []byte) uint16 {
	var crc uint16
	for _, c := range b {
		crc ^= uint16(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
