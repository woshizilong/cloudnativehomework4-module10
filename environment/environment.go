package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	BuildInfo string // 编译阶段注入的信息。格式：应用名:版本号[换行]Commit:git版本sha1码
	Version   string // 版本
	CommitID  string // 编译阶段注入的git版本sha1码

	StartTime time.Time // 应用启动时间
	Hostname  string    // 节点主机名
	ExecENV   string    // 当前的执行环境
)

const (
	envKey         = "HWENV"      // 标明当前执行环境的环境变量KEY
	envProduction  = "Production" // 生产集群环境
	envDevelopment = "Develop"    // 开发集群环境
	envLocalhost   = "Localhost"  // 单机环境
)

func init() {
	var err error
	// 记录启动时间
	StartTime = time.Now()
	// 获取节点名
	Hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
	// 获取当前环境
	ExecENV = os.Getenv(envKey)
	if !IsProduction() && !IsDevelopment() && !IsLocalhost() {
		// 默认环境为单机环境
		// ExecENV = envLocalhost
	}
}

// 是否处于生产集群环境
func IsProduction() bool {
	return ExecENV == envProduction
}

// 是否处于开发集群环境
func IsDevelopment() bool {
	return ExecENV == envDevelopment
}

// 是否处于单机环境
func IsLocalhost() bool {
	return ExecENV == envLocalhost
}

// 程序信息
func AppInfo() string {
	return fmt.Sprintf("%s\n", BuildInfo) +
		fmt.Sprintf("Hostname:\t%s\n", Hostname) +
		fmt.Sprintf("Environment:\t%s\n", ExecENV) +
		fmt.Sprintf("Start time:\t%s\n", StartTime.Format("2006-01-02 15:04:05")) +
		fmt.Sprintf("Running time:\t%s\n", time.Since(StartTime))
}

func AppName() string {
	path, _ := os.Executable()
	_, exec := filepath.Split(path)
	return exec
}
