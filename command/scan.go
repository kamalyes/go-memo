/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-23 09:18:27
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 09:18:27
 * @FilePath: \go-memo\command\scan.go
 * @Description: 键枚举命令与通配符匹配，支持 SCAN / KEYS
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// defaultScanCount SCAN 默认单批返回键数
const defaultScanCount = 10

func handleScan(st store.Store, args []string) resp.Value {
	cursor, pattern, count, err := parseScanArgs(args)
	if err != nil {
		return resp.ErrorString(err.Error())
	}

	all := st.Keys()
	sort.Strings(all)

	start := int(cursor)
	if start > len(all) {
		start = 0
	}

	var matched []string
	for i := start; i < len(all) && i < start+count; i++ {
		if pattern == "" || globMatch(pattern, all[i]) {
			matched = append(matched, all[i])
		}
	}

	next := start + count
	if next >= len(all) {
		next = 0
	}

	keys := make([]resp.Value, 0, len(matched))
	for _, k := range matched {
		keys = append(keys, resp.BulkString(k))
	}
	return resp.Array(resp.BulkString(strconv.Itoa(next)), resp.Array(keys...))
}

func handleKeys(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("keys"))
	}
	pattern := args[1]
	var keys []resp.Value
	for _, k := range st.Keys() {
		if globMatch(pattern, k) {
			keys = append(keys, resp.BulkString(k))
		}
	}
	return resp.Array(keys...)
}

// parseScanArgs 解析 SCAN cursor [MATCH pattern] [COUNT count] 参数
func parseScanArgs(args []string) (cursor uint64, pattern string, count int, err error) {
	if len(args) < 2 {
		return 0, "", 0, errors.New(errWrongArgs("scan"))
	}
	cursor, err = strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return 0, "", 0, errors.New(msgNotInteger)
	}
	count = defaultScanCount
	for i := 2; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "MATCH":
			if i+1 >= len(args) {
				return 0, "", 0, errors.New(errWrongArgs("scan"))
			}
			pattern = args[i+1]
			i++
		case "COUNT":
			if i+1 >= len(args) {
				return 0, "", 0, errors.New(errWrongArgs("scan"))
			}
			n, perr := strconv.Atoi(args[i+1])
			if perr != nil || n <= 0 {
				return 0, "", 0, errors.New(msgNotInteger)
			}
			count = n
			i++
		default:
			return 0, "", 0, errors.New(msgSyntaxError)
		}
	}
	return cursor, pattern, count, nil
}

// globMatch 实现 Redis 风格通配符匹配，支持 * ? [...] [^...] 与反斜杠转义
func globMatch(pattern, s string) bool {
	return globMatchRunes([]rune(pattern), []rune(s))
}

func globMatchRunes(p, s []rune) bool {
	for len(p) > 0 {
		switch p[0] {
		case '*':
			for len(p) > 0 && p[0] == '*' {
				p = p[1:]
			}
			if len(p) == 0 {
				return true
			}
			for i := 0; i <= len(s); i++ {
				if globMatchRunes(p, s[i:]) {
					return true
				}
			}
			return false
		case '?':
			if len(s) == 0 {
				return false
			}
			p = p[1:]
			s = s[1:]
		case '[':
			if len(s) == 0 {
				return false
			}
			closed, rest, matched := matchClass(p, s[0])
			if !closed {
				if s[0] != '[' {
					return false
				}
				p = p[1:]
				s = s[1:]
				continue
			}
			if !matched {
				return false
			}
			p = rest
			s = s[1:]
		case '\\':
			if len(p) >= 2 {
				p = p[1:]
			}
			fallthrough
		default:
			if len(s) == 0 || s[0] != p[0] {
				return false
			}
			p = p[1:]
			s = s[1:]
		}
	}
	return len(s) == 0
}

// matchClass 解析 [...] 字符类，返回是否闭合、剩余模式与首字符命中结果
func matchClass(p []rune, c rune) (closed bool, rest []rune, matched bool) {
	i := 1
	negate := false
	if i < len(p) && (p[i] == '^' || p[i] == '!') {
		negate = true
		i++
	}
	for ; i < len(p); i++ {
		if p[i] == ']' {
			closed = true
			rest = p[i+1:]
			break
		}
		if i+2 < len(p) && p[i+1] == '-' && p[i+2] != ']' {
			if c >= p[i] && c <= p[i+2] {
				matched = true
			}
			i += 2
		} else if c == p[i] {
			matched = true
		}
	}
	if negate {
		matched = !matched
	}
	return closed, rest, matched
}
