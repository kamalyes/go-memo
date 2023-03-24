/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-21 09:12:35
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-21 09:12:35
 * @FilePath: \go-memo\persistence\aof.go
 * @Description: 仅追加日志（AOF），顺序追加命令并周期刷盘
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package persistence

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/kamalyes/go-memo/resp"
)

// syncInterval AOF 刷盘周期
const syncInterval = time.Second

// AOF 仅追加日志，写路径顺序追加，后台协程按秒刷盘
type AOF struct {
	mu   sync.Mutex // 保护文件写入的互斥锁
	f    *os.File   // 底层追加日志文件
	path string     // 日志文件路径

	stopOnce sync.Once     // 停止信号只关闭一次
	stop     chan struct{} // 停止后台刷盘信号
	done     chan struct{} // 后台刷盘退出信号
}

// OpenAOF 打开或创建 AOF 文件并启动后台刷盘
func OpenAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	a := &AOF{f: f, path: path, stop: make(chan struct{}), done: make(chan struct{})}
	go a.syncLoop()
	return a, nil
}

// Append 以 RESP 数组形式追加一条命令
func (a *AOF) Append(args []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err := a.f.Write(resp.MarshalCommand(args))
	return err
}

// Sync 立即刷盘
func (a *AOF) Sync() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.f.Sync()
}

// Close 停止后台刷盘并最终同步关闭文件
func (a *AOF) Close() error {
	a.stopOnce.Do(func() { close(a.stop) })
	<-a.done

	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.f.Sync(); err != nil {
		return err
	}
	return a.f.Close()
}

func (a *AOF) syncLoop() {
	defer close(a.done)
	t := time.NewTicker(syncInterval)
	defer t.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-t.C:
			_ = a.Sync()
		}
	}
}

// Replay 顺序读取命令流并逐条回放，用于启动重建或快照装载
func Replay(path string, dispatch func([]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		args, err := resp.ReadCommand(br)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := dispatch(args); err != nil {
			return err
		}
	}
}
