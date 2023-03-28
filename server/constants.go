/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-19 09:11:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-19 09:11:05
 * @FilePath: \go-memo\server\constants.go
 * @Description: 服务默认配置常量定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package server

// DefaultAddr 默认监听地址
const DefaultAddr = ":7399"

// DBCount 逻辑库数量，与 Redis 默认 16 个库对齐
const DBCount = 16

// 服务器级命令固定的响应消息
const (
	msgOK         = "OK"
	msgIndexRange = "ERR DB index is out of range"
	msgEmptyCmd   = "ERR empty command"
	msgNotInteger = "ERR value is not an integer or out of range"
)

// errWrongArgs 构造参数数量错误
func errWrongArgs(name string) string {
	return "ERR wrong number of arguments for '" + name + "' command"
}

// errUnknownSubcommand 构造未知子命令错误
func errUnknownSubcommand(name string) string {
	return "ERR unknown subcommand '" + name + "'"
}
