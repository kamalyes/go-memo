/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-05-08 09:36:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-08 09:36:00
 * @FilePath: \go-memo\command\helpers.go
 * @Description: 命令层公共辅助，供各类型命令复用
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"sort"
	"strconv"

	"github.com/kamalyes/go-memo/resp"
)

// bulkArray 将字符串切片转换为批量字符串数组响应
func bulkArray(vals []string) resp.Value {
	out := make([]resp.Value, len(vals))
	for i, v := range vals {
		out[i] = resp.BulkString(v)
	}
	return resp.Array(out...)
}

// sortedKeys 返回映射键名升序切片，保证输出确定性
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedSame 返回字符串切片升序副本
func sortedSame(vals []string) []string {
	out := append([]string(nil), vals...)
	sort.Strings(out)
	return out
}

// formatScore 将分数格式化为最短可往返的十进制表示
func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'g', -1, 64)
}