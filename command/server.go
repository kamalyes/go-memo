/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:58:02
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-23 10:05:00
 * @FilePath: \go-memo\command\server.go
 * @Description: 服务级命令实现，连接态命令已迁移至 server 层
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

import (
	"github.com/kamalyes/go-memo/resp"
	"github.com/kamalyes/go-memo/store"
)

func handlePing(_ store.Store, _ []string) resp.Value {
	return resp.SimpleString(msgPong)
}
