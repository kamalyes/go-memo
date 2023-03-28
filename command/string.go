/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:32:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 11:02:00
 * @FilePath: \go-memo\command\string.go
 * @Description: 字符串类型命令实现
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

func handleGet(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("get"))
	}
	v, ok := st.Get(args[1])
	if !ok {
		return resp.NullBulk()
	}
	return resp.BulkString(v)
}

// handleSet 写入键值，支持 EX / PX / EXAT / PXAT / NX / XX 选项
func handleSet(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("set"))
	}
	key, value := args[1], args[2]
	var nx, xx bool
	var expireAt int64

	for i := 3; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "NX":
			nx = true
		case "XX":
			xx = true
		case "EX":
			i++
			if i >= len(args) {
				return resp.ErrorString(msgSyntaxError)
			}
			sec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil || sec <= 0 {
				return resp.ErrorString(msgNotInteger)
			}
			expireAt = time.Now().UnixMilli() + sec*1000
		case "PX":
			i++
			if i >= len(args) {
				return resp.ErrorString(msgSyntaxError)
			}
			ms, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil || ms <= 0 {
				return resp.ErrorString(msgNotInteger)
			}
			expireAt = time.Now().UnixMilli() + ms
		case "EXAT":
			i++
			if i >= len(args) {
				return resp.ErrorString(msgSyntaxError)
			}
			sec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil || sec <= 0 {
				return resp.ErrorString(msgNotInteger)
			}
			expireAt = sec * 1000
		case "PXAT":
			i++
			if i >= len(args) {
				return resp.ErrorString(msgSyntaxError)
			}
			ms, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil || ms <= 0 {
				return resp.ErrorString(msgNotInteger)
			}
			expireAt = ms
		default:
			return resp.ErrorString(msgSyntaxError)
		}
	}

	if nx {
		if !st.SetNX(key, value) {
			return resp.NullBulk()
		}
	} else if xx {
		if !st.SetXX(key, value) {
			return resp.NullBulk()
		}
	} else {
		st.Set(key, value)
	}
	if expireAt != 0 {
		st.Expire(key, expireAt)
	}
	return resp.SimpleString(msgOK)
}

func handleIncrBy(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("incrby"))
	}
	delta, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	return doIncrBy(st, args[1], delta)
}

func handleIncr(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("incr"))
	}
	return doIncrBy(st, args[1], 1)
}

func handleDecr(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("decr"))
	}
	return doIncrBy(st, args[1], -1)
}

func handleDecrBy(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("decrby"))
	}
	delta, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	return doIncrBy(st, args[1], -delta)
}

// doIncrBy 统一执行整数递增并包装标准错误
func doIncrBy(st store.Store, key string, delta int64) resp.Value {
	n, err := st.IncrBy(key, delta)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	return resp.Integer(n)
}

func handleAppend(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("append"))
	}
	cur, _ := st.Get(args[1])
	val := cur + args[2]
	st.Set(args[1], val)
	return resp.Integer(int64(len(val)))
}

func handleStrLen(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("strlen"))
	}
	v, ok := st.Get(args[1])
	if !ok {
		return resp.Integer(0)
	}
	return resp.Integer(int64(len(v)))
}

func handleMGet(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("mget"))
	}
	out := make([]resp.Value, 0, len(args)-1)
	for _, k := range args[1:] {
		if v, ok := st.Get(k); ok {
			out = append(out, resp.BulkString(v))
		} else {
			out = append(out, resp.NullBulk())
		}
	}
	return resp.Array(out...)
}

func handleMSet(st store.Store, args []string) resp.Value {
	if len(args) < 3 || len(args)%2 == 0 {
		return resp.ErrorString(errWrongArgs("mset"))
	}
	for i := 1; i < len(args); i += 2 {
		st.Set(args[i], args[i+1])
	}
	return resp.SimpleString(msgOK)
}

func handleSetEX(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("setex"))
	}
	sec, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil || sec <= 0 {
		return resp.ErrorString(msgNotInteger)
	}
	st.Set(args[1], args[3])
	st.Expire(args[1], time.Now().UnixMilli()+sec*1000)
	return resp.SimpleString(msgOK)
}

func handleSetNX(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("setnx"))
	}
	if st.SetNX(args[1], args[2]) {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}
