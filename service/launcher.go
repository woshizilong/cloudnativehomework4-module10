package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kabacloud/cloudnativehomework4-module10/environment"
	"github.com/kabacloud/cloudnativehomework4-module10/logger"
	"github.com/kabacloud/cloudnativehomework4-module10/middleware"
	"github.com/prometheus/client_golang/prometheus"
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

	// 定义指标
	// 创建一个自定义的注册表
	registry := prometheus.NewRegistry()
	// 可选: 添加 process 和 Go 运行时指标到我们自定义的注册表中
	// registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	// registry.MustRegister(prometheus.NewGoCollector())

	// 创建一个简单呃 gauge 指标。
	// gauge 类型的指标值是可以上升或下降，所以 gauge 指标对象暴露了 Set()、Inc()、Dec()、Add(float64) 和 Sub(float64) 这些函数来更改指标值。
	workerTimestamp := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_current_time",
		Help: "The current temperature in degrees Celsius.",
	})
	workerTimestamp.SetToCurrentTime()

	// 设置 gague 的值为 当前时间
	workerTimestamp.SetToCurrentTime()

	// counter 指标只能随着时间的推移而不断增加，所以我们不能为其设置一个指定的值或者减少指标值，所以该对象下面只有 Inc() 和 Add(float64) 两个函数
	totalRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_requests_total",
		Help: "The total number of handled HTTP requests.",
	})
	totalRequests.Add(999)

	// Histograms 直方图指标比 counter 和 gauge 都要复杂，因为需要配置把观测值归入的 bucket 的数量，以及每个 bucket 的上边界。
	// Prometheus 中的直方图是累积的，所以每一个后续的 bucket 都包含前一个 bucket 的观察计数，所有 bucket 的下限都从 0 开始的，
	// 所以我们不需要明确配置每个 bucket 的下限，只需要配置上限即可。
	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "worker_request_duration_seconds",
		Help: "A histogram of the HTTP request durations in seconds.",
		// Bucket 配置：第一个 bucket 包括所有在 0.05s 内完成的请求，最后一个包括所有在10s内完成的请求。
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})
	// 这里和前面不一样的地方在于除了指定指标名称和帮助信息之外，还需要配置 Buckets。如果我们手动去枚举所有的 bucket 可能很繁琐，
	// 所以 Go 客户端库为为我们提供了一些辅助函数可以帮助我们生成线性或者指数增长的 bucket，比如 prometheus.LinearBuckets() 和 prometheus.ExponentialBuckets() 函数。
	// 直方图会自动对数值的分布进行分类和计数，所以它只有一个 Observe(float64) 方法，每当你在代码中处理要跟踪的数据时，就会调用这个方法。
	// 例如，如果你刚刚处理了一个 HTTP 请求，花了 0.42 秒，则可以使用下面的代码来跟踪。
	requestDurations.Observe(0.99)
	// 由于跟踪持续时间是直方图的一个常见用例，Go 客户端库就提供了辅助函数，用于对代码的某些部分进行计时，然后自动观察所产生的持续时间，将其转化为直方图，如下代码所示：
	/*
		```go
		// 启动一个计时器
		timer := prometheus.NewTimer(requestDurations)

		// [...在应用中处理请求...]

		// 停止计时器并观察其持续时间，将其放进 requestDurations 的直方图指标中去
		timer.ObserveDuration()
		```
	*/
	// 每个配置的存储桶最终作为一个带有 _bucket 后缀的计数器时间序列，使用 le（小于或等于） 标签指示该存储桶的上限，
	// 具有上限的隐式存储桶 +Inf 也暴露于比最大配置的存储桶边界花费更长的时间的请求，还包括使用后缀 _sum 累积总和和计数 _count 的指标，
	// 这些时间序列中的每一个在概念上都是一个 counter 计数器（只能上升的单个值），只是它们是作为直方图的一部分创建的。
	// 结果如：
	// 	http_request_duration_seconds_bucket{le="0.05"} 4599
	//  http_request_duration_seconds_bucket{le="0.1"} 24128
	//  http_request_duration_seconds_bucket{le="0.25"} 45311
	//  http_request_duration_seconds_bucket{le="0.5"} 59983
	//  http_request_duration_seconds_bucket{le="1"} 60345
	//  http_request_duration_seconds_bucket{le="2.5"} 114003
	//  http_request_duration_seconds_bucket{le="5"} 201325
	//  http_request_duration_seconds_bucket{le="+Inf"} 227420
	//  http_request_duration_seconds_sum 88364.234
	//  http_request_duration_seconds_count 227420

	// Summaries 创建和使用摘要与直方图非常类似，只是我们需要指定要跟踪的 quantiles 分位数值，而不需要处理 bucket 桶，
	// 比如我们想要跟踪 HTTP 请求延迟的第 50、90 和 99 个百分位数，那么我们可以创建这样的一个摘要对象：
	appRequestDurations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "app_request_duration_seconds",
		Help: "A summary of the HTTP request durations in seconds.",
		Objectives: map[float64]float64{
			0.5:  0.05,  // 第50个百分位数，最大绝对误差为0.05。
			0.9:  0.01,  // 第90个百分位数，最大绝对误差为0.01。
			0.99: 0.001, // 第90个百分位数，最大绝对误差为0.001。
		},
	},
	)
	// 这里和前面不一样的地方在于使用 prometheus.NewSummary() 函数初始化摘要指标对象的时候，需要通过 prometheus.SummaryOpts{} 对象的 Objectives 属性指定想要跟踪的分位数值。
	// 同样摘要指标对象创建后，跟踪持续时间的方式和直方图是完全一样的，使用一个 Observe(float64) 函数即可：
	appRequestDurations.Observe(0.77)

	// 创建带 worker 和 app 标签的 gauge 指标对象
	appWorker := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_temperature_celsius",
			Help: "The current temperature in degrees Celsius.",
		},
		// 指定标签名称
		[]string{"worker", "service"},
	)
	// 针对不同标签值设置不同的指标值
	appWorker.WithLabelValues(environment.Hostname, "log").Set(1)
	appWorker.WithLabelValues(environment.Hostname, "trace").Set(2)
	// 注意：当使用带有标签维度的指标时，任何标签组合的时间序列只有在该标签组合被访问过至少一次后才会出现在 /metrics 输出中，
	// 这对我们在 PromQL 查询的时候会产生一些问题，因为它希望某些时间序列一直存在，我们可以在程序第一次启动时，将所有重要的标签组合预先初始化为默认值。

	// 使用我们自定义的注册表注册自定义指标
	registry.MustRegister(workerTimestamp)
	registry.MustRegister(totalRequests)
	registry.MustRegister(requestDurations)
	registry.MustRegister(appRequestDurations)
	registry.MustRegister(appWorker)

	// 定义路由
	// k8s关于健康检查API的说明 https://kubernetes.io/zh/docs/reference/using-api/health-checks/
	http.Handle("/healthz", middleware.ResponseLog(http.HandlerFunc(healthHandler))) // 健康检查
	http.Handle("/livez", middleware.ResponseLog(http.HandlerFunc(healthHandler)))   // 健康检查
	http.Handle("/readyz", middleware.ResponseLog(http.HandlerFunc(readyHandler)))   // 就绪检查
	// k8s指标监控
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	// 服务功能API
	http.Handle("/info", middleware.RequestHeader(middleware.ResponseLog(http.HandlerFunc(infoHandler)))) // 基本功能
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
