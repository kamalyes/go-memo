/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-22 11:22:37
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-22 11:22:37
 * @FilePath: \go-memo\replication\replication_test.go
 * @Description: 扇出、全量同步与选举状态机测试
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package replication

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/kamalyes/go-memo/command"
	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func TestMasterFanout(t *testing.T) {
	m := NewMaster()
	var b1, b2 bytes.Buffer
	r1 := m.Attach(&b1)
	m.Attach(&b2)

	if n := m.Fanout([]string{"SET", "k", "v"}); n != 2 {
		t.Fatalf("Fanout = %d, want 2", n)
	}
	want := string(resp.MarshalCommand([]string{"SET", "k", "v"}))
	if b1.String() != want || b2.String() != want {
		t.Fatalf("replica bytes mismatch: %q / %q", b1.String(), b2.String())
	}

	m.Detach(r1)
	if n := m.Fanout([]string{"SET", "k2", "v2"}); n != 1 {
		t.Fatalf("Fanout after detach = %d, want 1", n)
	}
}

func TestReplicationFullSync(t *testing.T) {
	m := NewMaster()
	src := store.New()
	src.Set("foo", "bar")
	src.Set("cnt", "5")

	dst := store.New()
	reg := command.NewRegistry()

	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		_ = m.ServeReplica(server, src.Snapshot)
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Sync(client, func(args []string) error {
			reg.Dispatch(dst, args)
			return nil
		})
	}()

	if v := waitGet(t, dst, "foo"); v != "bar" {
		t.Fatalf("synced foo = %q, want bar", v)
	}
	if v := waitGet(t, dst, "cnt"); v != "5" {
		t.Fatalf("synced cnt = %q, want 5", v)
	}

	m.Fanout([]string{"SET", "late", "value"})
	if v := waitGet(t, dst, "late"); v != "value" {
		t.Fatalf("incremental late = %q, want value", v)
	}

	_ = client.Close()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Sync did not return after client close")
	}
}

func TestElection(t *testing.T) {
	e := NewElection("A", 1, 3*time.Second)
	e.Heartbeat("B", 2, 5)
	if e.Leader() != "B" || e.Epoch() != 5 {
		t.Fatalf("leader = %s epoch %d, want B 5", e.Leader(), e.Epoch())
	}

	base := time.Now().Add(10 * time.Second)
	for i := 0; i < missThreshold; i++ {
		e.Tick(base.Add(time.Duration(i) * time.Second))
	}
	if e.Leader() != "A" {
		t.Fatalf("leader after timeout = %s, want A", e.Leader())
	}
	if e.Epoch() != 6 {
		t.Fatalf("epoch after promote = %d, want 6", e.Epoch())
	}
}

func waitGet(t *testing.T, st store.Store, key string) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if v, ok := st.Get(key); ok {
			return v
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("key %q not synced", key)
	return ""
}
