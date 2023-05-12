/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-09 09:55:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-09 09:55:00
 * @FilePath: \go-memo\command\hash.go
 * @Description: 哈希类型命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// parseHashFields 解析 HSET/HMSET 的 field value 序列，返回字段映射与是否合法
func parseHashFields(args []string, cmd string) (map[string]string, bool) {
	if len(args) < 4 || (len(args)-2)%2 != 0 {
		return nil, false
	}
	fields := make(map[string]string, (len(args)-2)/2)
	for i := 2; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}
	return fields, true
}

func handleHSet(st store.Store, args []string) resp.Value {
	fields, ok := parseHashFields(args, "hset")
	if !ok {
		return resp.ErrorString(errWrongArgs("hset"))
	}
	n, err := st.HSetMap(args[1], fields)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(int64(n))
}

func handleHMSet(st store.Store, args []string) resp.Value {
	fields, ok := parseHashFields(args, "hmset")
	if !ok {
		return resp.ErrorString(errWrongArgs("hmset"))
	}
	if _, err := st.HSetMap(args[1], fields); err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.SimpleString(msgOK)
}

func handleHSetNX(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("hsetnx"))
	}
	ok, err := st.HSetNX(args[1], args[2], args[3])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return boolInt(ok)
}

func handleHGet(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("hget"))
	}
	v, ok, err := st.HGet(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if !ok {
		return resp.NullBulk()
	}
	return resp.BulkString(v)
}

func handleHMGet(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("hmget"))
	}
	vals, oks, err := st.HMGet(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	out := make([]resp.Value, len(vals))
	for i := range vals {
		if oks[i] {
			out[i] = resp.BulkString(vals[i])
		} else {
			out[i] = resp.NullBulk()
		}
	}
	return resp.Array(out...)
}

func handleHGetAll(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("hgetall"))
	}
	m, err := st.HGetAll(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	keys := sortedKeys(m)
	out := make([]resp.Value, 0, len(keys)*2)
	for _, f := range keys {
		out = append(out, resp.BulkString(f), resp.BulkString(m[f]))
	}
	return resp.Array(out...)
}

func handleHDel(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("hdel"))
	}
	n, err := st.HDel(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(int64(n))
}

func handleHLen(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("hlen"))
	}
	n, err := st.HLen(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleHExists(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("hexists"))
	}
	ok, err := st.HExists(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return boolInt(ok)
}

func handleHKeys(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("hkeys"))
	}
	keys, err := st.HKeys(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(keys))
}

func handleHVals(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("hvals"))
	}
	vals, err := st.HVals(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(vals))
}

func handleHIncrBy(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("hincrby"))
	}
	delta, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	n, err := st.HIncrBy(args[1], args[2], delta)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}
