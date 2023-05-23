/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:08:31
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-05-05 09:25:00
 * @FilePath: \go-memo\store\interfaces.go
 * @Description: 键值存储契约定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// Store 键值存储契约，所有后端（内存、持久化）均面向该接口
type Store interface {
	// Type 返回键的当前类型，键不存在或已过期返回 TypeNone
	Type(key string) ValueType

	// Get 返回字符串键对应值，键不存在返回空与 false，类型不符返回 ErrWrongType
	Get(key string) (string, bool, error)

	// Set 写入字符串键值，不带过期时间
	Set(key string, value string)

	// SetNX 仅在键不存在时写入字符串，返回是否写入成功
	SetNX(key string, value string) bool

	// SetXX 仅在键已存在时写入字符串，返回是否写入成功
	SetXX(key string, value string) bool

	// IncrBy 将字符串键值按整数递增，键不存在时视为 0，非整数值返回错误
	IncrBy(key string, delta int64) (int64, error)

	// Del 删除一个或多个键，返回实际删除的存活键数量
	Del(keys ...string) int

	// Rename 将源键值移动到目标键，源键不存在返回 false
	Rename(src, dst string) bool

	// Expire 为已存在的键设置过期时间，键不存在或已过期返回 false
	Expire(key string, expireAt int64) bool

	// Persist 移除键的过期时间，键不存在或未设置过期返回 false
	Persist(key string) bool

	// TTL 返回剩余存活毫秒数，语义见 TTLNotExist / TTLNoExpire
	TTL(key string) int64

	// Keys 返回所有存活键名，过期键被过滤
	Keys() []string

	// Exists 判断键是否存活
	Exists(key string) bool

	// Keyspace 返回存活键数量与带过期时间的键数量
	Keyspace() (keys int64, expires int64)

	// Snapshot 返回全量存活键值快照，供持久化与复制采集
	Snapshot() []Record

	// Flush 清空存储内全部键
	Flush()

	// UsedMemory 返回数据集估算内存占用字节
	UsedMemory() int64

	// 列表操作
	LPush(key string, values ...string) (int64, error)
	RPush(key string, values ...string) (int64, error)
	LPushX(key string, values ...string) (int64, error)
	RPushX(key string, values ...string) (int64, error)
	LRange(key string, start, stop int) ([]string, error)
	LLen(key string) (int64, error)
	LPop(key string, count int) ([]string, bool, error)
	RPop(key string, count int) ([]string, bool, error)
	LRem(key string, count int, value string) (int64, error)
	LSet(key string, index int, value string) error
	LIndex(key string, index int) (string, bool, error)
	LInsert(key string, before bool, pivot, value string) (int64, error)
	LTrim(key string, start, stop int) error

	// 哈希操作
	HSet(key, field, value string) (int, error)
	HSetMap(key string, fields map[string]string) (int, error)
	HSetNX(key, field, value string) (bool, error)
	HGet(key, field string) (string, bool, error)
	HMGet(key string, fields ...string) ([]string, []bool, error)
	HGetAll(key string) (map[string]string, error)
	HDel(key string, fields ...string) (int, error)
	HLen(key string) (int64, error)
	HExists(key, field string) (bool, error)
	HKeys(key string) ([]string, error)
	HVals(key string) ([]string, error)
	HIncrBy(key, field string, delta int64) (int64, error)

	// 集合操作
	SAdd(key string, members ...string) (int, error)
	SRem(key string, members ...string) (int, error)
	SMembers(key string) ([]string, error)
	SIsMember(key, member string) (bool, error)
	SCard(key string) (int64, error)
	SPop(key string, count int) ([]string, bool, error)
	SRandMember(key string, count int) ([]string, error)
	SMove(src, dst, member string) (bool, error)
	SInter(keys ...string) ([]string, error)
	SUnion(keys ...string) ([]string, error)
	SDiff(keys ...string) ([]string, error)
	SInterStore(dst string, keys ...string) (int64, error)
	SUnionStore(dst string, keys ...string) (int64, error)
	SDiffStore(dst string, keys ...string) (int64, error)

	// 有序集合操作
	ZAdd(key string, pairs ...ZPair) (int, error)
	ZCard(key string) (int64, error)
	ZScore(key, member string) (float64, bool, error)
	ZIncrBy(key, member string, delta float64) (float64, error)
	ZRem(key string, members ...string) (int, error)
	ZRange(key string, start, stop int) ([]ZPair, error)
	ZRevRange(key string, start, stop int) ([]ZPair, error)
	ZRangeByScore(key, min, max string, offset, count int) ([]ZPair, error)
	ZRevRangeByScore(key, max, min string, offset, count int) ([]ZPair, error)
	ZRank(key, member string) (int64, bool, error)
	ZRevRank(key, member string) (int64, bool, error)
	ZCount(key, min, max string) (int64, error)
	ZRemRangeByRank(key string, start, stop int) (int64, error)
	ZRemRangeByScore(key, min, max string) (int64, error)
}
