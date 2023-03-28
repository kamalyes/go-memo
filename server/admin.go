/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-23 11:20:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 11:20:00
 * @FilePath: \go-memo\server\admin.go
 * @Description: 服务器级与连接态命令实现
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kamalyes/go-memo/command"
	"github.com/kamalyes/go-memo/resp"
)

// cmdSelect 切换当前连接的逻辑库
func (s *Server) cmdSelect(ci *connInfo, args []string) resp.Value {
	if len(args) != 2 {
		return resp.ErrorString(errWrongArgs("select"))
	}
	idx, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.ErrorString(msgNotInteger)
	}
	if idx < 0 || idx >= DBCount {
		return resp.ErrorString(msgIndexRange)
	}
	ci.db = int(idx)
	return resp.SimpleString(msgOK)
}

// cmdQuit 标记连接退出，写回 OK 后由读写循环关闭
func (s *Server) cmdQuit(ci *connInfo, _ []string) resp.Value {
	ci.quit = true
	return resp.SimpleString(msgOK)
}

// cmdFlushDB 清空当前逻辑库
func (s *Server) cmdFlushDB(ci *connInfo, _ []string) resp.Value {
	s.dbs[ci.db].Flush()
	return resp.SimpleString(msgOK)
}

// cmdFlushAll 清空全部逻辑库
func (s *Server) cmdFlushAll(_ []string) resp.Value {
	for i := range s.dbs {
		s.dbs[i].Flush()
	}
	return resp.SimpleString(msgOK)
}

// cmdClient 处理 CLIENT 子命令，LIST 返回当前连接列表
func (s *Server) cmdClient(ci *connInfo, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorString(errWrongArgs("client"))
	}
	switch strings.ToUpper(args[1]) {
	case "LIST":
		return resp.BulkString(s.clientList())
	case "SETNAME":
		if len(args) != 3 {
			return resp.ErrorString(errWrongArgs("client|setname"))
		}
		ci.name = args[2]
		return resp.SimpleString(msgOK)
	case "GETNAME":
		return resp.BulkString(ci.name)
	case "ID":
		return resp.Integer(ci.id)
	case "SETINFO":
		return resp.SimpleString(msgOK)
	}
	return resp.ErrorString(errUnknownSubcommand(args[1]))
}

// clientList 序列化当前活动连接列表，供 CLIENT LIST 返回
func (s *Server) clientList() string {
	s.mu.Lock()
	conns := make([]*connInfo, 0, len(s.conns))
	for _, ci := range s.conns {
		conns = append(conns, ci)
	}
	s.mu.Unlock()
	sort.Slice(conns, func(i, j int) bool { return conns[i].id < conns[j].id })

	now := time.Now()
	var b strings.Builder
	for _, ci := range conns {
		age := int64(now.Sub(ci.created).Seconds())
		idle := int64(now.Sub(ci.lastCmd).Seconds())
		fmt.Fprintf(&b, "id=%d addr=%s name=%s db=%d age=%d idle=%d flags=N sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 argv-mem=0 obl=0 oll=0 omem=0 tot-mem=0 events=r cmd=client user=default redir=-1 resp=2\r\n",
			ci.id, ci.addr, ci.name, ci.db, age, idle)
	}
	return b.String()
}

// configItem 单个可查询配置项
type configItem struct {
	name  string
	value string
}

// configItems 返回服务器支持的配置项集合
func (s *Server) configItems() []configItem {
	return []configItem{
		{name: "databases", value: strconv.Itoa(DBCount)},
		{name: "maxmemory", value: "0"},
		{name: "maxclients", value: "10000"},
		{name: "appendonly", value: "no"},
		{name: "save", value: ""},
		{name: "dir", value: "."},
		{name: "timeout", value: "0"},
		{name: "tcp-keepalive", value: "300"},
		{name: "daemonize", value: "no"},
	}
}

// cmdConfig 处理 CONFIG GET/SET，仅支持查询已知配置项
func (s *Server) cmdConfig(args []string) resp.Value {
	if len(args) < 3 {
		return resp.ErrorString(errWrongArgs("config"))
	}
	switch strings.ToUpper(args[1]) {
	case "GET":
		pattern := args[2]
		items := s.configItems()
		var out []resp.Value
		for _, it := range items {
			if pattern == "*" || strings.EqualFold(it.name, pattern) {
				out = append(out, resp.BulkString(it.name), resp.BulkString(it.value))
			}
		}
		return resp.Array(out...)
	case "SET":
		return resp.SimpleString(msgOK)
	}
	return resp.ErrorString(errUnknownSubcommand(args[1]))
}

// cmdCommand 处理 COMMAND 子命令，COUNT 返回已注册命令数
func (s *Server) cmdCommand(args []string) resp.Value {
	if len(args) >= 2 && strings.EqualFold(args[1], "COUNT") {
		return resp.Integer(int64(s.reg.Count()))
	}
	return resp.Array()
}

// cmdHello 处理 HELLO 握手，返回服务器能力，RESP2 下以扁平键值数组表达
func (s *Server) cmdHello(ci *connInfo, _ []string) resp.Value {
	out := []resp.Value{
		resp.BulkString("server"), resp.BulkString("go-memo"),
		resp.BulkString("version"), resp.BulkString(command.Version),
		resp.BulkString("proto"), resp.BulkString("2"),
		resp.BulkString("id"), resp.Integer(ci.id),
		resp.BulkString("mode"), resp.BulkString("standalone"),
		resp.BulkString("role"), resp.BulkString("master"),
		resp.BulkString("modules"), resp.Array(),
	}
	return resp.Array(out...)
}

// cmdCluster 处理 CLUSTER 子命令，单机模式返回禁用状态
func (s *Server) cmdCluster(args []string) resp.Value {
	if len(args) >= 2 {
		switch strings.ToUpper(args[1]) {
		case "INFO":
			return resp.BulkString("cluster_state:ok\r\ncluster_enabled:0\r\n")
		case "NODES":
			return resp.BulkString("")
		}
	}
	return resp.ErrorString("ERR This instance has cluster support disabled")
}

// cmdInfo 输出服务器信息文本，不含 Redis 伪装字段，仅保留 memo 自身字段
func (s *Server) cmdInfo(_ []string) resp.Value {
	return resp.BulkString(s.infoText())
}

// infoText 组装 INFO 各段，覆盖 Server / Clients / Memory / Keyspace 等
func (s *Server) infoText() string {
	stats := s.reg.Stats()
	secs := int64(time.Since(stats.StartTime()).Seconds())

	var b strings.Builder
	b.WriteString("# Server\r\n")
	fmt.Fprintf(&b, "memo_version:%s\r\n", command.Version)
	fmt.Fprintf(&b, "run_id:%s\r\n", stats.RunID())
	fmt.Fprintf(&b, "tcp_port:%d\r\n", stats.Port())
	fmt.Fprintf(&b, "uptime_in_seconds:%d\r\n", secs)
	fmt.Fprintf(&b, "uptime_in_days:%d\r\n", secs/86400)
	fmt.Fprintf(&b, "databases:%d\r\n", DBCount)

	b.WriteString("\r\n# Clients\r\n")
	fmt.Fprintf(&b, "connected_clients:%d\r\n", stats.ConnCount())

	b.WriteString("\r\n# Memory\r\n")
	var used int64
	for i := range s.dbs {
		used += s.dbs[i].UsedMemory()
	}
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
	for i := range s.dbs {
		keys, expires := s.dbs[i].Keyspace()
		if keys == 0 {
			continue
		}
		fmt.Fprintf(&b, "db%d:keys=%d,expires=%d,avg_ttl=0\r\n", i, keys, expires)
	}
	return b.String()
}

// humanBytes 将字节数格式化为人类可读单位
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
