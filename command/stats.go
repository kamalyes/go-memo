/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-23 10:22:16
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 10:22:16
 * @FilePath: \go-memo\command\stats.go
 * @Description: 服务运行期统计状态，由 server 层更新并由 INFO 命令消费
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"time"
)

// Stats 服务运行期统计状态，集中维护连接、端口与命令计数
type Stats struct {
	startTime time.Time // 服务启动时刻
	runID     string    // 本次运行唯一标识

	connCount atomic.Int64  // 当前连接数
	totalConn atomic.Int64  // 累计接收连接数
	tcpPort   atomic.Int64  // 实际监听端口
	cmdCount  atomic.Uint64 // 累计处理命令数
}

// NewStats 创建统计状态，初始化启动时刻与运行标识
func NewStats() *Stats {
	return &Stats{startTime: time.Now(), runID: newRunID()}
}

// newRunID 生成 40 位十六进制运行标识
func newRunID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("0", 40)
	}
	return hex.EncodeToString(b)
}

// IncrCmd 累计一次命令处理
func (s *Stats) IncrCmd() { s.cmdCount.Add(1) }

// IncrConn 累计一条新连接建立
func (s *Stats) IncrConn() { s.connCount.Add(1) }

// DecrConn 累计一条连接关闭
func (s *Stats) DecrConn() { s.connCount.Add(-1) }

// IncrTotalConn 累计接收连接总数
func (s *Stats) IncrTotalConn() { s.totalConn.Add(1) }

// SetPort 设置实际监听端口
func (s *Stats) SetPort(port int64) { s.tcpPort.Store(port) }

// StartTime 返回服务启动时刻
func (s *Stats) StartTime() time.Time { return s.startTime }

// RunID 返回本次运行唯一标识
func (s *Stats) RunID() string { return s.runID }

// ConnCount 返回当前连接数
func (s *Stats) ConnCount() int64 { return s.connCount.Load() }

// TotalConn 返回累计接收连接数
func (s *Stats) TotalConn() int64 { return s.totalConn.Load() }

// Port 返回实际监听端口
func (s *Stats) Port() int64 { return s.tcpPort.Load() }

// CmdCount 返回累计处理命令数
func (s *Stats) CmdCount() uint64 { return s.cmdCount.Load() }
