/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-22 09:05:16
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-22 09:05:16
 * @FilePath: \go-memo\replication\constants.go
 * @Description: 复制协议与选举参数常量定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package replication

import "time"

// 复制协议常量
const (
	cmdSync = "SYNC"

	// endOfSnapshot 全量快照结束标记，空数组响应
	endOfSnapshot = "*-1\r\n"
)

// 选举参数
const (
	missThreshold = 3 // 连续 missThreshold 次心跳未达判死
)

// DefaultHeartbeat 默认心跳间隔
const DefaultHeartbeat = 3 * time.Second
