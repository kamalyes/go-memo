/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-11 09:52:30
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-11 09:52:30
 * @FilePath: \go-memo\command\zset.go
 * @Description: 有序集合类型命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"strconv"
	"strings"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handleZAdd(st store.Store, args []string) resp.Value {
	if len(args) < 4 {
		return resp.ErrorString(errWrongArgs("zadd"))
	}
	key := args[1]
	nx, xx, gt, lt, ch, incr := false, false, false, false, false, false
	i := 2
optLoop:
	for i < len(args) {
		switch strings.ToUpper(args[i]) {
		case "NX":
			nx = true
		case "XX":
			xx = true
		case "GT":
			gt = true
		case "LT":
			lt = true
		case "CH":
			ch = true
		case "INCR":
			incr = true
		default:
			break optLoop
		}
		i++
	}
	if nx && xx {
		return resp.ErrorString("ERR XX and NX options at the same time are not compatible")
	}
	if nx && (gt || lt) {
		return resp.ErrorString("ERR GT, LT, and/or NX options at the same time are not compatible")
	}
	rest := args[i:]
	if len(rest) == 0 || len(rest)%2 != 0 {
		return resp.ErrorString(msgSyntaxError)
	}
	if incr && len(rest) != 2 {
		return resp.ErrorString("ERR INCR option supports a single increment-element pair")
	}
	pairs := make([]store.ZPair, 0, len(rest)/2)
	for j := 0; j < len(rest); j += 2 {
		score, err := strconv.ParseFloat(rest[j], 64)
		if err != nil {
			return resp.ErrorString(msgNotFloat)
		}
		pairs = append(pairs, store.ZPair{Member: rest[j+1], Score: score})
	}
	if incr {
		n, err := st.ZIncrBy(key, pairs[0].Member, pairs[0].Score)
		if err != nil {
			return resp.ErrorString(err.Error())
		}
		return resp.BulkString(formatScore(n))
	}
	added, changed, toAdd, err := filterZAddPairs(st, key, pairs, nx, xx, gt, lt)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if len(toAdd) > 0 {
		if _, err := st.ZAdd(key, toAdd...); err != nil {
			return resp.ErrorString(err.Error())
		}
	}
	if ch {
		return resp.Integer(int64(changed))
	}
	return resp.Integer(int64(added))
}

// filterZAddPairs 按 NX/XX/GT/LT 过滤待写入成员，返回新增数、更改数与保留列表
func filterZAddPairs(st store.Store, key string, pairs []store.ZPair, nx, xx, gt, lt bool) (int, int, []store.ZPair, error) {
	added, changed := 0, 0
	toAdd := make([]store.ZPair, 0, len(pairs))
	for _, p := range pairs {
		cur, exists, err := st.ZScore(key, p.Member)
		if err != nil {
			return 0, 0, nil, err
		}
		if !exists {
			if xx {
				continue
			}
			toAdd = append(toAdd, p)
			added++
			continue
		}
		if nx {
			continue
		}
		if gt && p.Score <= cur {
			continue
		}
		if lt && p.Score >= cur {
			continue
		}
		toAdd = append(toAdd, p)
		if p.Score != cur {
			changed++
		}
	}
	return added, changed, toAdd, nil
}

func handleZCard(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("zcard"))
	}
	n, err := st.ZCard(args[1])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleZScore(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("zscore"))
	}
	score, ok, err := st.ZScore(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if !ok {
		return resp.NullBulk()
	}
	return resp.BulkString(formatScore(score))
}

func handleZIncrBy(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("zincrby"))
	}
	delta, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.ErrorString(msgNotFloat)
	}
	n, err := st.ZIncrBy(args[1], args[3], delta)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.BulkString(formatScore(n))
}

func handleZRem(st store.Store, args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("zrem"))
	}
	n, err := st.ZRem(args[1], args[2:]...)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(int64(n))
}

func handleZRange(st store.Store, args []string) resp.Value {
	if len(args) < 4 || len(args) > 5 {
		return resp.ErrorString(errWrongArgs("zrange"))
	}
	start, err1 := strconv.Atoi(args[2])
	stop, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.ErrorString(msgNotInteger)
	}
	withScores := len(args) == 5 && strings.EqualFold(args[4], "WITHSCORES")
	pairs, err := st.ZRange(args[1], start, stop)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return zsetResult(pairs, withScores)
}

func handleZRevRange(st store.Store, args []string) resp.Value {
	if len(args) < 4 || len(args) > 5 {
		return resp.ErrorString(errWrongArgs("zrevrange"))
	}
	start, err1 := strconv.Atoi(args[2])
	stop, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.ErrorString(msgNotInteger)
	}
	withScores := len(args) == 5 && strings.EqualFold(args[4], "WITHSCORES")
	pairs, err := st.ZRevRange(args[1], start, stop)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return zsetResult(pairs, withScores)
}

func handleZRangeByScore(st store.Store, args []string) resp.Value {
	if len(args) < 4 {
		return resp.ErrorString(errWrongArgs("zrangebyscore"))
	}
	withScores, offset, count, ok := parseZRangeOptions(args, 4)
	if !ok {
		return resp.ErrorString(msgSyntaxError)
	}
	pairs, err := st.ZRangeByScore(args[1], args[2], args[3], offset, count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return zsetResult(pairs, withScores)
}

func handleZRevRangeByScore(st store.Store, args []string) resp.Value {
	if len(args) < 4 {
		return resp.ErrorString(errWrongArgs("zrevrangebyscore"))
	}
	withScores, offset, count, ok := parseZRangeOptions(args, 4)
	if !ok {
		return resp.ErrorString(msgSyntaxError)
	}
	pairs, err := st.ZRevRangeByScore(args[1], args[2], args[3], offset, count)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return zsetResult(pairs, withScores)
}

// parseZRangeOptions 解析 WITHSCORES 与 LIMIT offset count 选项
func parseZRangeOptions(args []string, idx int) (withScores bool, offset, count int, ok bool) {
	count = -1
	for idx < len(args) {
		switch strings.ToUpper(args[idx]) {
		case "WITHSCORES":
			withScores = true
			idx++
		case "LIMIT":
			if idx+2 >= len(args) {
				return false, 0, 0, false
			}
			o, err1 := strconv.Atoi(args[idx+1])
			c, err2 := strconv.Atoi(args[idx+2])
			if err1 != nil || err2 != nil {
				return false, 0, 0, false
			}
			offset, count = o, c
			idx += 3
		default:
			return false, 0, 0, false
		}
	}
	return withScores, offset, count, true
}

func handleZRank(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("zrank"))
	}
	rank, ok, err := st.ZRank(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if !ok {
		return resp.NullBulk()
	}
	return resp.Integer(rank)
}

func handleZRevRank(st store.Store, args []string) resp.Value {
	if len(args) != 3 {
		return resp.ErrorString(errWrongArgs("zrevrank"))
	}
	rank, ok, err := st.ZRevRank(args[1], args[2])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	if !ok {
		return resp.NullBulk()
	}
	return resp.Integer(rank)
}

func handleZCount(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("zcount"))
	}
	n, err := st.ZCount(args[1], args[2], args[3])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleZRemRangeByRank(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("zremrangebyrank"))
	}
	start, err1 := strconv.Atoi(args[2])
	stop, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		return resp.ErrorString(msgNotInteger)
	}
	n, err := st.ZRemRangeByRank(args[1], start, stop)
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

func handleZRemRangeByScore(st store.Store, args []string) resp.Value {
	if len(args) != 4 {
		return resp.ErrorString(errWrongArgs("zremrangebyscore"))
	}
	n, err := st.ZRemRangeByScore(args[1], args[2], args[3])
	if err != nil {
		return resp.ErrorString(err.Error())
	}
	return resp.Integer(n)
}

// zsetResult 将有序集合成员对组装为数组响应，withScores 时成员与分数交替
func zsetResult(pairs []store.ZPair, withScores bool) resp.Value {
	if withScores {
		out := make([]resp.Value, 0, len(pairs)*2)
		for _, p := range pairs {
			out = append(out, resp.BulkString(p.Member), resp.BulkString(formatScore(p.Score)))
		}
		return resp.Array(out...)
	}
	out := make([]resp.Value, len(pairs))
	for i, p := range pairs {
		out[i] = resp.BulkString(p.Member)
	}
	return resp.Array(out...)
}
