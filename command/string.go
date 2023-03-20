/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:32:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 09:32:11
 * @FilePath: \go-memo\command\string.go
 * @Description: 字符串类型命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"

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

func handleSet(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("set"))
	}
	st.Set(args[1], args[2])
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
	n, err := st.IncrBy(args[1], delta)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	return resp.Integer(n)
}