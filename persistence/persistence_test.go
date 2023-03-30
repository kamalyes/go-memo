/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-21 11:18:52
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-21 11:18:52
 * @FilePath: \go-memo\persistence\persistence_test.go
 * @Description: AOF 追加回放与 RDB 快照装载测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package persistence

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kamalyes/go-memo/command"
	"github.com/kamalyes/go-memo/store"
)

func replayInto(st store.Store, path string) error {
	reg := command.NewRegistry()
	return Replay(path, func(args []string) error {
		reg.Dispatch(st, args)
		return nil
	})
}

func TestAOFAppendReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")
	aof, err := OpenAOF(path)
	if err != nil {
		t.Fatalf("OpenAOF: %v", err)
	}
	for _, cmd := range [][]string{
		{"SET", "foo", "bar"},
		{"SET", "num", "100"},
		{"INCRBY", "num", "5"},
		{"PEXPIRE", "foo", "60000"},
	} {
		if err := aof.Append(cmd); err != nil {
			t.Fatalf("Append %v: %v", cmd, err)
		}
	}
	if err := aof.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	st := store.New()
	if err := replayInto(st, path); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if v, ok := st.Get("foo"); !ok || v != "bar" {
		t.Fatalf("foo = %q, %v", v, ok)
	}
	if v, ok := st.Get("num"); !ok || v != "105" {
		t.Fatalf("num = %q, %v", v, ok)
	}
	if ttl := st.TTL("foo"); ttl <= 0 {
		t.Fatalf("foo TTL = %d, want > 0", ttl)
	}
}

func TestReplayMissingFile(t *testing.T) {
	if err := replayInto(store.New(), filepath.Join(t.TempDir(), "none.aof")); err != nil {
		t.Fatalf("Replay missing file should be nil, got %v", err)
	}
}

func TestRDBSaveLoad(t *testing.T) {
	src := store.New()
	src.Set("foo", "bar")
	src.Set("cnt", "3")
	src.Expire("foo", time.Now().UnixMilli()+60_000)
	src.Expire("gone", time.Now().UnixMilli()-1)

	path := filepath.Join(t.TempDir(), "dump.rdb")
	n, err := Save(src, path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if n != 2 {
		t.Fatalf("snapshot records = %d, want 2 (expired key filtered)", n)
	}

	dst := store.New()
	if err := Load(path, func(args []string) error {
		command.NewRegistry().Dispatch(dst, args)
		return nil
	}); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if v, ok := dst.Get("foo"); !ok || v != "bar" {
		t.Fatalf("foo = %q, %v", v, ok)
	}
	if v, ok := dst.Get("cnt"); !ok || v != "3" {
		t.Fatalf("cnt = %q, %v", v, ok)
	}
	if _, ok := dst.Get("gone"); ok {
		t.Fatal("expired key should not be restored")
	}
	if ttl := dst.TTL("foo"); ttl <= 0 {
		t.Fatalf("foo TTL = %d, want > 0", ttl)
	}
}
