/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-22 09:35:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-22 09:35:58
 * @FilePath: \go-memo\replication\election.go
 * @Description: 零依赖领袖选举，静态优先级加 epoch 心跳抢占防脑裂
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package replication

import (
	"sync"
	"time"
)

// Election 单节点视角的领袖选举状态机
type Election struct {
	mu       sync.Mutex    // 保护以下字段的互斥锁
	self     string        // 本节点标识
	priority int           // 静态优先级，越大越优先
	interval time.Duration // 心跳超时判定间隔

	leader   string    // 当前认定领袖
	epoch    uint64    // 领袖任期号，单调递增防脑裂
	lastBeat time.Time // 最近一次心跳时刻
	miss     int       // 连续未达心跳次数
}

// NewElection 创建选举状态机，priority 越大越优先
func NewElection(self string, priority int, interval time.Duration) *Election {
	return &Election{
		self:     self,
		priority: priority,
		interval: interval,
		leader:   self,
		epoch:    0,
		lastBeat: time.Now(),
	}
}

// Heartbeat 记录一次候选心跳，epoch 更大者或同 epoch 更高优先级者晋升
func (e *Election) Heartbeat(id string, priority int, epoch uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastBeat = time.Now()
	e.miss = 0

	if id == e.leader {
		e.epoch = epoch
		return
	}
	if epoch > e.epoch || (epoch == e.epoch && priority > e.priority) {
		e.leader = id
		e.epoch = epoch
	}
}

// Tick 周期推进，领袖心跳超时后本节点抢占
func (e *Election) Tick(now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader == e.self {
		return
	}
	if now.Sub(e.lastBeat) < e.interval {
		return
	}
	e.miss++
	if e.miss < missThreshold {
		return
	}
	e.leader = e.self
	e.epoch++
	e.miss = 0
	e.lastBeat = now
}

// Leader 返回当前认定领袖
func (e *Election) Leader() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.leader
}

// Epoch 返回当前领袖 epoch
func (e *Election) Epoch() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.epoch
}
