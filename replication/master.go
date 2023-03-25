/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-22 09:18:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-22 09:18:33
 * @FilePath: \go-memo\replication\master.go
 * @Description: 复制主节点，登记副本并扇出命令流
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package replication

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/kamalyes/go-memo/persistence"
	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// errBadSync 复制握手请求非法
var errBadSync = errors.New("replication: invalid sync request")

// Replica 已登记的副本连接句柄
type Replica struct {
	mu sync.Mutex    // 保护写入缓冲的互斥锁
	bw *bufio.Writer // 指向副本连接的写缓冲
}

// Master 复制主节点，维护副本集合并扇出命令流
type Master struct {
	mu       sync.Mutex            // 保护副本集合的互斥锁
	replicas map[*Replica]struct{} // 已登记副本集合
}

// NewMaster 创建主节点
func NewMaster() *Master {
	return &Master{replicas: make(map[*Replica]struct{})}
}

// Attach 将外部写入器登记为副本，返回句柄
func (m *Master) Attach(w io.Writer) *Replica {
	return m.attach(bufio.NewWriter(w))
}

// Detach 移除副本
func (m *Master) Detach(r *Replica) {
	m.mu.Lock()
	delete(m.replicas, r)
	m.mu.Unlock()
}

// Fanout 将命令扇出到所有副本，返回写入成功的副本数
func (m *Master) Fanout(args []string) int {
	data := resp.MarshalCommand(args)
	m.mu.Lock()
	reps := make([]*Replica, 0, len(m.replicas))
	for r := range m.replicas {
		reps = append(reps, r)
	}
	m.mu.Unlock()

	n := 0
	for _, r := range reps {
		r.mu.Lock()
		_, werr := r.bw.Write(data)
		if werr == nil {
			werr = r.bw.Flush()
		}
		r.mu.Unlock()
		if werr != nil {
			m.Detach(r)
			continue
		}
		n++
	}
	return n
}

// ServeReplica 处理一条副本连接：SYNC 全量快照后登记进入增量扇出
func (m *Master) ServeReplica(conn net.Conn, snapshot func() []store.Record) error {
	defer conn.Close()

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	args, err := resp.ReadCommand(br)
	if err != nil {
		return err
	}
	if len(args) == 0 || !strings.EqualFold(args[0], cmdSync) {
		return errBadSync
	}

	for _, cmd := range persistence.SnapshotCommands(snapshot()) {
		if _, err := bw.Write(resp.MarshalCommand(cmd)); err != nil {
			return err
		}
	}
	if _, err := bw.WriteString(endOfSnapshot); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	r := m.attach(bw)
	defer m.Detach(r)

	for {
		if _, err := br.ReadByte(); err != nil {
			return err
		}
	}
}

func (m *Master) attach(bw *bufio.Writer) *Replica {
	r := &Replica{bw: bw}
	m.mu.Lock()
	m.replicas[r] = struct{}{}
	m.mu.Unlock()
	return r
}
