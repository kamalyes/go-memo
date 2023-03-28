/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:35:26
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 09:35:26
 * @FilePath: \go-memo\command\key.go
 * @Description: 键管理与过期时间命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"
	"time"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handleDel(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("del"))
	}
	return resp.Integer(int64(st.Del(args[1:]...)))
}

func handleExpire(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("expire"))
	}
	sec, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if sec <= 0 {
		return resp.Integer(int64(st.Del(args[1])))
	}
	return boolInt(st.Expire(args[1], time.Now().UnixMilli()+sec*1000))
}

func handlePExpire(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("pexpire"))
	}
	ms, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if ms <= 0 {
		return resp.Integer(int64(st.Del(args[1])))
	}
	return boolInt(st.Expire(args[1], time.Now().UnixMilli()+ms))
}

func handleTTL(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("ttl"))
	}
	rem := st.TTL(args[1])
	switch rem {
	case store.TTLNotExist:
		return resp.Integer(-2)
	case store.TTLNoExpire:
		return resp.Integer(-1)
	}
	return resp.Integer((rem + 500) / 1000)
}

func handlePTTL(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("pttl"))
	}
	return resp.Integer(st.TTL(args[1]))
}

// handleType 返回键值类型，仅支持字符串与不存在
func handleType(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("type"))
	}
	if st.Exists(args[1]) {
		return resp.SimpleString("string")
	}
	return resp.SimpleString("none")
}

// handleDBSize 返回当前库存活键数量
func handleDBSize(st store.Store, args []string) resp.Value {
	if len(args) != 1 {
		return resp.ErrorString(errWrongArgs("dbsize"))
	}
	keys, _ := st.Keyspace()
	return resp.Integer(keys)
}

// handleExists 返回存活键数量
func handleExists(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("exists"))
	}
	n := 0
	for _, k := range args[1:] {
		if st.Exists(k) {
			n++
		}
	}
	return resp.Integer(int64(n))
}

// handlePersist 移除键的过期时间
func handlePersist(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("persist"))
	}
	if st.Persist(args[1]) {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

// handleRename 将源键改名到目标键
func handleRename(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("rename"))
	}
	if !st.Rename(args[1], args[2]) {
		return resp.ErrorString("ERR no such key")
	}
	return resp.SimpleString(msgOK)
}

func boolInt(b bool) resp.Value {
	if b {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}
