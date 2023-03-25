/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-21 09:35:27
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-21 09:35:27
 * @FilePath: \go-memo\persistence\rdb.go
 * @Description: 全量快照（RDB），读锁采集后写临时文件原子重命名
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package persistence

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// Save 将存储全量快照写入 path，返回快照键数量
func Save(m *store.Memory, path string) (int, error) {
	records := m.Snapshot()
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return 0, err
	}
	defer func() { _ = os.Remove(tmp) }()

	w := bufio.NewWriter(f)
	for _, cmd := range SnapshotCommands(records) {
		if err := writeCommand(w, cmd); err != nil {
			_ = f.Close()
			return 0, err
		}
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return 0, err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return 0, err
	}
	if err := f.Close(); err != nil {
		return 0, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return 0, err
	}
	return len(records), nil
}

// Load 回放快照文件重建存储
func Load(path string, dispatch func([]string) error) error {
	return Replay(path, dispatch)
}

// SnapshotCommands 将快照记录转换为可回放的 SET/PEXPIRE 命令序列
func SnapshotCommands(records []store.Record) [][]string {
	now := time.Now().UnixMilli()
	cmds := make([][]string, 0, len(records)*2)
	for _, r := range records {
		cmds = append(cmds, []string{"SET", r.Key, r.Value})
		if r.ExpireAt != 0 {
			if rem := r.ExpireAt - now; rem > 0 {
				cmds = append(cmds, []string{"PEXPIRE", r.Key, strconv.FormatInt(rem, 10)})
			}
		}
	}
	return cmds
}

func writeCommand(w *bufio.Writer, args []string) error {
	_, err := w.Write(resp.MarshalCommand(args))
	return err
}
