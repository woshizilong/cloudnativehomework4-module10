package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kabacloud/cloudnativehomework4-module10/cmd"
	_ "go.uber.org/automaxprocs"
)

// 容器内根据获得的CPU数自动设定GOMAXPROCS
func main() {
	// 设定日志时间精确到微秒
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	// 准备上下文，用于监听到终止信号时能结束子命令
	ctxMain, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听信号，命令被终止时，能进行后续收尾工作。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGKILL, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM) // signal.Notify(quit) 监听所有信号
	go func() {
		for sig := range quit {
			// 通知各模块进行退出处理
			fmt.Printf("\n=========\n收到系统信号 %s\n", sig)
			cancel()
		}
	}()

	cmd.Execute(ctxMain)
}
