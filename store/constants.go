/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:02:13
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-17 09:02:13
 * @FilePath: \go-memo\store\constants.go
 * @Description: 内存存储分片计数与过期键存活性常量定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// shardCount 分片数量，通过哈希分散到不同锁降低竞争
const shardCount = 256

// 存活时间查询（TTL）返回值语义
const (
	TTLNotExist = int64(-2) // 键不存在
	TTLNoExpire = int64(-1) // 键存在但永不过期
)

// ValueType 键值数据类型，TypeNone 表示键不存在
type ValueType uint8

// 键值数据类型枚举
const (
	TypeNone   ValueType = iota // 键不存在
	TypeString                  // 字符串
	TypeList                    // 列表
	TypeHash                    // 哈希
	TypeSet                     // 集合
	TypeZSet                    // 有序集合
)

// String 返回类型对应的 Redis 类型名，TypeNone 返回 none
func (t ValueType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeHash:
		return "hash"
	case TypeSet:
		return "set"
	case TypeZSet:
		return "zset"
	default:
		return "none"
	}
}
