# 全局参数
ARG BASE_IMAGE

# 分阶段生成镜像的第一阶段:编译
FROM golang:1.16-alpine AS builder

# 设定编译环境的环境变量和工作目录
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/homework

# 拷贝编译用文件
COPY go.mod go.sum ./
COPY . .
# 编译应用（通过 --build-arg 引入的参数每次使用前都要ARG一次，否则第二个使用参数的命令是接不到的。）
ARG BUILD_INFO
RUN go build -mod=vendor -o httpserver -v -ldflags '-X "github.com/kabacloud/cloudnativehomework4-module10/environment.BuildInfo='"${BUILD_INFO}"'" '

# 分阶段生成镜像的第二阶段:设定
FROM ${BASE_IMAGE}

WORKDIR /homework

# 拷贝第一阶段编译好的APP到工作目录
COPY --from=builder /go/homework/httpserver httpserver
COPY config.yaml config.yaml

# 设定APP启动命令和命令参数
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/httpserver", "serve"]