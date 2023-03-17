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