/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-17 09:08:31
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-17 09:08:31
 * @FilePath: \go-memo\store\interfaces.go
 * @Description: 键值存储契约定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package store

// Store 键值存储契约，所有后端（内存、持久化）均面向该接口
type Store interface {
	// Get 返回键对应值，键不存在或已过期时第二个返回值为 false
	Get(key string) (string, bool)

	// Set 写入键值，不带过期时间
	Set(key string, value string)

	// SetNX 仅在键不存在时写入，返回是否写入成功
	SetNX(key string, value string) bool

	// SetXX 仅在键已存在时写入，返回是否写入成功
	SetXX(key string, value string) bool

	// IncrBy 将键值按整数递增，键不存在时视为 0，非整数值返回错误
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

	// Flush 清空存储内全部键
	Flush()

	// UsedMemory 返回数据集估算内存占用字节
	UsedMemory() int64
}
