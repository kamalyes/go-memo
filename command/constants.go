/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-03-18 09:05:20
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-03-18 09:05:20
 * @FilePath: \go-memo\command\constants.go
 * @Description: 命令版本号与错误消息常量定义
 *
 * Copyright (c) 2023 by kamalyes, All Rights Reserved.
 */

package command

// Version 服务版本号，供 INFO 命令返回
const Version = "0.1.0"

// 固定错误消息
const (
	msgNotInteger  = "ERR value is not an integer or out of range"
	msgEmptyCmd    = "ERR empty command"
	msgSyntaxError = "ERR syntax error"
	msgPong        = "PONG"
	msgOK          = "OK"
)

// errUnknownCommand 构造未识别命令错误
func errUnknownCommand(name string) string {
	return "ERR unknown command '" + name + "'"
}

// errWrongArgs 构造参数数量错误
func errWrongArgs(name string) string {
	return "ERR wrong number of arguments for '" + name + "' command"
}
