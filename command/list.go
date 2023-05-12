/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-08 09:31:12
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-08 09:31:12
 * @FilePath: \go-memo\command\list.go
 * @Description: 列表类型命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handleLPush(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("lpush"))
	}
	n, err := st.LPush(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleRPush(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("rpush"))
	}
	n, err := st.RPush(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleLPushX(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("lpushx"))
	}
	n, err := st.LPushX(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleRPushX(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("rpushx"))
	}
	n, err := st.RPushX(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleLRange(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("lrange"))
	}
	start, err1 := strconv.Atoi(args[2])
	stop, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.ErrorString(msgNotInteger)
	}
	vals, err := st.LRange(args[1], start, stop)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(vals)
}

func handleLLen(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("llen"))
	}
	n, err := st.LLen(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleLPop(st store.Store, args []string) resp.Value {
	count, ok := parseOptionalCount(args, "lpop")
	if !ok {
		return resp.ErrorString(errWrongArgs("lpop"))
	}
	vals, existed, err := st.LPop(args[1], count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return popResult(args, vals, existed)
}

func handleRPop(st store.Store, args []string) resp.Value {
	count, ok := parseOptionalCount(args, "rpop")
	if !ok {
		return resp.ErrorString(errWrongArgs("rpop"))
	}
	vals, existed, err := st.RPop(args[1], count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return popResult(args, vals, existed)
}

// parseOptionalCount 解析可选的 count 参数，缺失时返回 1
func parseOptionalCount(args []string, cmd string) (int, bool) {
	if len(args) == 2 {
		return 1, true
	}
	if len(args) != 3 {
		return 0, false
	}
	n, err := strconv.Atoi(args[2])
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// popResult 将弹出结果按是否携带 count 参数组装响应
func popResult(args []string, vals []string, existed bool) resp.Value {
	if len(args) == 2 {
		if !existed || len(vals) == 0 {
			return resp.NullBulk()
		}
		return resp.BulkString(vals[0])
	}
	if !existed || len(vals) == 0 {
		return resp.NullArray()
	}
	return bulkArray(vals)
}

func handleLRem(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("lrem"))
	}
	count, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	n, err := st.LRem(args[1], count, args[3])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleLSet(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("lset"))
	}
	index, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if err := st.LSet(args[1], index, args[3]); err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.SimpleString(msgOK)
}

func handleLIndex(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("lindex"))
	}
	index, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	v, ok, err := st.LIndex(args[1], index)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if !ok {
		return resp.NullBulk()
	}
	return resp.BulkString(v)
}

func handleLInsert(st store.Store, args []string) resp.Value {
	if len(args) != 5 {
		return resp.ErrorString(errWrongArgs("linsert"))
	}
	before := false
	switch args[2] {
	case "BEFORE":
		before = true
	case "AFTER":
		before = false
	default:
		return resp.ErrorString(msgSyntaxError)
	}
	n, err := st.LInsert(args[1], before, args[3], args[4])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleLTrim(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("ltrim"))
	}
	start, err1 := strconv.Atoi(args[2])
	stop, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if err := st.LTrim(args[1], start, stop); err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.SimpleString(msgOK)
}