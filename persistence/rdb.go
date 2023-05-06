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
	"sort"
	"strconv"
	"time"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

// Save 将存储全量快照写入 path，返回快照键数量
func Save(m *store.Memory, path string) (int, error) {
	records := m.Snapshot()
	if err := writeStream(path, SnapshotCommands(records)); err != nil {
		return 0, err
	}
	return len(records), nil
}

// SaveDBs 将多个逻辑库全量快照写入 path，库间以 SELECT 分隔，返回总键数量
func SaveDBs(dbs []store.Store, path string) (int, error) {
	var cmds [][]string
	total := 0
	for i, db := range dbs {
		records := db.Snapshot()
		if len(records) == 0 {
			continue
		}
		total += len(records)
		cmds = append(cmds, []string{"SELECT", strconv.Itoa(i)})
		cmds = append(cmds, SnapshotCommands(records)...)
	}
	if err := writeStream(path, cmds); err != nil {
		return 0, err
	}
	return total, nil
}

// writeStream 将命令流写入临时文件后原子重命名到 path
func writeStream(path string, cmds [][]string) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp) }()

	w := bufio.NewWriter(f)
	for _, cmd := range cmds {
		if err := writeCommand(w, cmd); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load 回放快照文件重建存储
func Load(path string, dispatch func([]string) error) error {
	return Replay(path, dispatch)
}

// SnapshotCommands 将快照记录转换为可回放的命令序列，按类型发出对应写命令
func SnapshotCommands(records []store.Record) [][]string {
	now := time.Now().UnixMilli()
	cmds := make([][]string, 0, len(records)*2)
	for _, r := range records {
		cmds = append(cmds, encodeRecord(r)...)
		if r.ExpireAt != 0 {
			if rem := r.ExpireAt - now; rem > 0 {
				cmds = append(cmds, []string{"PEXPIRE", r.Key, strconv.FormatInt(rem, 10)})
			}
		}
	}
	return cmds
}

// encodeRecord 将单条快照记录编码为对应类型的写命令
func encodeRecord(r store.Record) [][]string {
	switch r.Typ {
	case store.TypeString:
		return [][]string{{"SET", r.Key, r.Str}}
	case store.TypeList:
		return [][]string{append([]string{"RPUSH", r.Key}, r.List...)}
	case store.TypeHash:
		args := []string{"HSET", r.Key}
		for _, f := range sortedKeys(r.Hash) {
			args = append(args, f, r.Hash[f])
		}
		return [][]string{args}
	case store.TypeSet:
		return [][]string{append([]string{"SADD", r.Key}, r.Set...)}
	case store.TypeZSet:
		args := []string{"ZADD", r.Key}
		for _, p := range r.ZSet {
			args = append(args, formatScore(p.Score), p.Member)
		}
		return [][]string{args}
	default:
		return nil
	}
}

// sortedKeys 返回映射键名升序切片，保证命令流确定性
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// formatScore 将分数格式化为最短可往返的十进制表示
func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'g', -1, 64)
}

func writeCommand(w *bufio.Writer, args []string) error {
	_, err := w.Write(resp.MarshalCommand(args))
	return err
}
