/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-22 09:30:21
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-22 09:30:21
 * @FilePath: \go-memo\replication\slave.go
 * @Description: 复制从节点，SYNC 全量同步与增量应用
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package replication

import (
	"bufio"
	"net"

	"github.com/kamalyes/go-memo/resp"
)

// Sync 作为从节点连接主节点：发送 SYNC，读取快照与增量命令并逐条应用
func Sync(conn net.Conn, dispatch func([]string) error) error {
	defer conn.Close()

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)
	if _, err := bw.Write(resp.MarshalCommand([]string{cmdSync})); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	for {
		args, err := resp.ReadCommand(br)
		if err != nil {
			return err
		}
		if len(args) == 0 {
			continue
		}
		if err := dispatch(args); err != nil {
			return err
		}
	}
}