package metrics

import (
	"github.com/kabacloud/cloudnativehomework4-module10/environment"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registry                 *prometheus.Registry
	httpserverSleepDurations prometheus.Histogram
)

func LoadRegistry() *prometheus.Registry {
	if registry == nil {

		// 定义指标
		// 创建一个自定义的注册表
		registry = prometheus.NewRegistry()
		// 可选: 添加 process 和 Go 运行时指标到我们自定义的注册表中
		// registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		// registry.MustRegister(prometheus.NewGoCollector())

		httpserverSleepDurations = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "httpserver_sleep_duration_seconds",
			Help: "A histogram of the HTTP request durations in seconds.",
			// Bucket 配置：第一个 bucket 包括所有在 0.5s 内完成的请求，最后一个包括所有在2.5s内完成的请求。
			Buckets: []float64{0.5, 1, 1.5, 2, 2.5},
		})

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
		registry.MustRegister(httpserverSleepDurations)
		registry.MustRegister(workerTimestamp)
		registry.MustRegister(totalRequests)
		registry.MustRegister(requestDurations)
		registry.MustRegister(appRequestDurations)
		registry.MustRegister(appWorker)
	}

	return registry
}

func RecordSleep(duration float64) {
	if registry != nil {
		httpserverSleepDurations.Observe(duration)
	}
}
