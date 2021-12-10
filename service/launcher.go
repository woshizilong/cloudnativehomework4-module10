package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/kabacloud/cloudnativehomework4-module10/environment"
	"github.com/kabacloud/cloudnativehomework4-module10/logger"
	"github.com/kabacloud/cloudnativehomework4-module10/metrics"
	"github.com/kabacloud/cloudnativehomework4-module10/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	isReady bool // 服务是否准备就绪
	log     *logger.LoggerProvider
)

func init() {
}

// 启动服务
func Start(ctxMain context.Context) error {
	var srv *http.Server

	log = logger.NewLogger("debug", "httpserver")

	// 根据环境区分的操作
	if environment.IsProduction() {
		log.Info("服务执行在生产环境下")
	} else {
		log.Info("服务执行在非生产环境下")
	}

	// 加载prometheus注册器
	r := metrics.LoadRegistry()

	// 定义路由
	// k8s关于健康检查API的说明 https://kubernetes.io/zh/docs/reference/using-api/health-checks/
	http.Handle("/healthz", middleware.ResponseLog(http.HandlerFunc(healthHandler))) // 健康检查
	http.Handle("/livez", middleware.ResponseLog(http.HandlerFunc(healthHandler)))   // 健康检查
	http.Handle("/readyz", middleware.ResponseLog(http.HandlerFunc(readyHandler)))   // 就绪检查
	// k8s指标监控
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{Registry: r}))
	// 服务功能API
	http.Handle("/info", middleware.RequestHeader(middleware.ResponseLog(http.HandlerFunc(infoHandler)))) // 基本功能
	http.Handle("/", middleware.RequestHeader(middleware.ResponseLog(http.HandlerFunc(infoHandler))))     // 基本功能
	// http.Handle("/giteataskrun", middleware.RequestHeader(middleware.ResponseLog(http.HandlerFunc(giteatask.GiteaWebhookHandler)))) // gitea webhook 触发 tekton 的 PipelineRun
	http.Handle("/run", middleware.RequestHeader(middleware.ResponseLog(http.HandlerFunc(runHandler)))) // 服务

	// 定义服务器
	srv = &http.Server{
		Addr:    ":8000",
		Handler: http.DefaultServeMux,
	}

	processed := make(chan struct{})

	// 通过传递的context监听退出信号
	go func() {
		log.Info("服务开始监听退出信号")
		<-ctxMain.Done()
		log.Info("服务监听到了退出信号")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		log.Info("服务停止接收新的请求")
		if err := srv.Shutdown(ctx); nil != err {
			log.FatalI("服务关闭失败", "error", err)
		}
		log.Info("服务已处理完现有请求")
		cleanup()
		log.Info("服务已完全关闭")
		close(processed)
	}()

	// 开始服务前的准备工作，比如读取配置、准备数据库连接等等工作
	go ready()
	err := srv.ListenAndServe()
	if http.ErrServerClosed != err {
		log.FatalI("server not gracefully shutdown", "error", err)
	}

	// 等待服务完全关闭
	<-processed

	return nil
}

// 开始前的准备工作
func ready() {
	log.Info("服务的准备工作开始进行")
	time.Sleep(10 * time.Second)
	isReady = true
	log.Info("服务的准备工作已完成")
}

// 停止服务前的收尾工作
func cleanup() {
	log.Info("服务的收尾工作开始进行")
	isReady = false
	time.Sleep(2 * time.Second)
	log.Info("服务的收尾工作已完成")
}

// 打印服务基本信息
func infoHandler(w http.ResponseWriter, r *http.Request) {
	// 添加 0-2 秒的随机延时
	rand.Seed(time.Now().UnixNano())
	min := 0.0
	max := 2.0
	duration := min + rand.Float64()*(max-min)
	time.Sleep(time.Second * time.Duration(duration))

	metrics.RecordSleep(duration)
	// 返回处理结果
	fmt.Fprint(w, environment.AppInfo())
}

// 健康检查用（k8s存活探针）
func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("healthHandler called")
	w.WriteHeader(http.StatusOK)
}

// 就绪检查用（k8s就绪探针）
func readyHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("readyHandler called")
	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 服务
func runHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("runHandler called")
	// 每次请求的statuscode只能写一次，向w的body写入时会默认尝试写入200。
	// 如果想自定义statuscode必须要在写入body前执行，否则就无效会报错“http: superfluous response.WriteHeader call from 你的代码”
	// 正确的设定顺序是 应答头（1） < 状态码（2） < 应答体（3）
	w.WriteHeader(http.StatusNonAuthoritativeInfo)
	fmt.Fprint(w, "有需求请联络管理员")
	// w.WriteHeader(http.StatusNonAuthoritativeInfo)
}
