package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/kabacloud/cloudnativehomework4-module10/environment"
)

func RequestHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 中间件的逻辑在这里实现,在执行传递进来的handler之前
		// [作业要求]Request中的Header要带入Response中
		for k, v := range r.Header {
			if environment.IsLocalhost() {
				log.Printf("%s:%s", k, v)
			}
			w.Header().Set(k, v[0]) // 用Postman测试自定义request header时，如果值是空的话，服务接收到的值是空串不是nil，所以直接用v[0]取值而没有判断nil。
		}
		// [作业要求]Response中带入环境变量VERSION的值
		envVersion := os.Getenv("VERSION")
		if len(envVersion) == 0 {
			envVersion = environment.Version
		}
		w.Header().Set("VERSION", envVersion)
		next.ServeHTTP(w, r)
		// 在handler执行之后的中间件逻辑
		// 如:执行时间的记录
	})
}
