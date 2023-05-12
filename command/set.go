/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-10 09:55:21
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-10 09:55:21
 * @FilePath: \go-memo\command\set.go
 * @Description: 集合类型命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handleSAdd(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("sadd"))
	}
	n, err := st.SAdd(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(int64(n))
}

func handleSRem(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("srem"))
	}
	n, err := st.SRem(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(int64(n))
}

func handleSMembers(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("smembers"))
	}
	members, err := st.SMembers(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(members))
}

func handleSIsMember(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("sismember"))
	}
	ok, err := st.SIsMember(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return boolInt(ok)
}

func handleSCard(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("scard"))
	}
	n, err := st.SCard(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleSPop(st store.Store, args []string) resp.Value {
	count, ok := parseOptionalCount(args, "spop")
	if !ok {
		return resp.ErrorString(errWrongArgs("spop"))
	}
	vals, existed, err := st.SPop(args[1], count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return popResult(args, vals, existed)
}

func handleSRandMember(st store.Store, args []string) resp.Value {
	if len(args) == 2 {
		vals, err := st.SRandMember(args[1], 1)
		if err != nil {
			return resp.ErrorString(err.Error())
		}
		if len(vals) == 0 {
			return resp.NullBulk()
		}
		return resp.BulkString(vals[0])
	}
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("srandmember"))
	}
	count, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	vals, err := st.SRandMember(args[1], count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(vals)
}

func handleSMove(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("smove"))
	}
	ok, err := st.SMove(args[1], args[2], args[3])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return boolInt(ok)
}

func handleSInter(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("sinter"))
	}
	result, err := st.SInter(args[1:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(result))
}

func handleSUnion(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("sunion"))
	}
	result, err := st.SUnion(args[1:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(result))
}

func handleSDiff(st store.Store, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("sdiff"))
	}
	result, err := st.SDiff(args[1:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return bulkArray(sortedSame(result))
}

func handleSInterStore(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("sinterstore"))
	}
	n, err := st.SInterStore(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleSUnionStore(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("sunionstore"))
	}
	n, err := st.SUnionStore(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleSDiffStore(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("sdiffstore"))
	}
	n, err := st.SDiffStore(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}