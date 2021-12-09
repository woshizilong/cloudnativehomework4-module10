package middleware

import (
	"log"
	"net"
	"net/http"
)

// 为了记录response的statusCode而定义的结构体
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

// 扩展了系统的http.ResponseWriter对象的WriteHeader方法，增加了记录statusCode的功能。
func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func ResponseLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 中间件的逻辑在这里实现,在执行传递进来的handler之前
		// 如:验证权限

		// [作业要求]取得IP后在标准输出中输出
		wRecorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		next.ServeHTTP(wRecorder, r)
		// 在handler执行之后的中间件逻辑
		// // [作业要求]Response中带入环境变量VERSION的值
		// envVersion := os.Getenv("VERSION")
		// if len(envVersion) == 0 {
		// 	envVersion = environment.Version
		// }
		// w.Header().Set("VERSION", envVersion)

		// [作业要求]取得IP后在标准输出中记录IP的返回状态码
		ip, _, _ := net.SplitHostPort(getIP(r))
		log.Printf("%s statusCode:%d url:%s", ip, wRecorder.Status, r.URL)
	})
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
