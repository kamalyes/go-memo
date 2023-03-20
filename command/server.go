/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:58:02
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 09:58:02
 * @FilePath: \go-memo\command\server.go
 * @Description: 服务级命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handlePing(_ store.Store, _ []string) resp.Value {
	return resp.SimpleString(msgPong)
}

func handleSelect(st store.Store, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("select"))
	}
	idx, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if idx != 0 {
		return resp.ErrorString(msgIndexRange)
	}
	return resp.SimpleString(msgOK)
}

func (r *Registry) handleInfo(st store.Store, _ []string) resp.Value {
	return resp.BulkString(infoText(st, r.stats))
}

// infoText 输出 Redis 兼容的 INFO 文本，含 Server / Clients / Memory / Stats 等标准段
func infoText(st store.Store, stats *Stats) string {
	keys, expires := st.Keyspace()
	used := st.UsedMemory()
	secs := int64(time.Since(stats.StartTime()).Seconds())

	var b strings.Builder
	b.WriteString("# Server\r\n")
	b.WriteString("redis_version:7.2.0\r\n")
	b.WriteString("redis_mode:standalone\r\n")
	fmt.Fprintf(&b, "memo_version:%s\r\n", Version)
	fmt.Fprintf(&b, "os:%s\r\n", runtime.GOOS)
	fmt.Fprintf(&b, "arch_bits:%d\r\n", strconv.IntSize)
	fmt.Fprintf(&b, "process_id:%d\r\n", os.Getpid())
	fmt.Fprintf(&b, "run_id:%s\r\n", stats.RunID())
	fmt.Fprintf(&b, "tcp_port:%d\r\n", stats.Port())
	fmt.Fprintf(&b, "uptime_in_seconds:%d\r\n", secs)
	fmt.Fprintf(&b, "uptime_in_days:%d\r\n", secs/86400)

	b.WriteString("\r\n# Clients\r\n")
	fmt.Fprintf(&b, "connected_clients:%d\r\n", stats.ConnCount())

	b.WriteString("\r\n# Memory\r\n")
	fmt.Fprintf(&b, "used_memory:%d\r\n", used)
	fmt.Fprintf(&b, "used_memory_human:%s\r\n", humanBytes(used))

	b.WriteString("\r\n# Stats\r\n")
	fmt.Fprintf(&b, "total_connections_received:%d\r\n", stats.TotalConn())
	fmt.Fprintf(&b, "total_commands_processed:%d\r\n", stats.CmdCount())

	b.WriteString("\r\n# Replication\r\n")
	b.WriteString("role:master\r\n")
	b.WriteString("connected_slaves:0\r\n")

	b.WriteString("\r\n# Cluster\r\n")
	b.WriteString("cluster_enabled:0\r\n")

	b.WriteString("\r\n# Keyspace\r\n")
	fmt.Fprintf(&b, "db0:keys=%d,expires=%d,avg_ttl=0\r\n", keys, expires)
	return b.String()
}

// humanBytes 将字节数格式化为人类可读单位，如 1.05M
func humanBytes(n int64) string {
	const unit = int64(1024)
	if n < unit {
		return strconv.FormatInt(n, 10) + "B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f%c", float64(n)/float64(div), "KMGTPE"[exp])
}
