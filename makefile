# 帮助信息用到的颜色
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

# Go 命令
GOROOT=$(shell go env GOROOT)
GOCMD=go
GOCLEAN=$(GOCMD) clean
# GOFMT=$(GOCMD) fmt
GOFMT=gofumpt -l -w
GOGENERATE=$(GOCMD) generate
GOBUILD=$(GOCMD) build -mod=vendor
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool

# 可执行文件名
APPNAME=httpserver
# 当前版本
APPVERSION=v0.4.0

# 提交标记SHA1码
COMMIT_SHA1=$(shell git rev-parse --short HEAD)

BUILD_INFO="$(APPNAME)@$(APPVERSION)-$(COMMIT_SHA1)"

# 编译平台
PLATFORM=

# -ldflags '-s -w' 
#    -s 去掉符号表，堆栈跟踪时不显示任何文件名和行号信息，跟C/C++的strip效果一样
#    -w 去掉DWARF调试信息,使程序不能gdb调试
LDFLAGs=

# 根据编译环境决定应用程序名
ifeq ($(OS),Windows_NT)
	PLATFORM=Windows
	BINARY_NAME=$(APPNAME)_windows_amd64.exe
	LDFLAGs=-ldflags "-X github.com/kabacloud/cloudnativehomework4-module10/environment.BuildInfo=$(BUILD_INFO) "
else
	ifeq ($(shell uname),Darwin)
		PLATFORM=MacOS
		BINARY_NAME=$(APPNAME)
	else
		PLATFORM=Unix-Like
		BINARY_NAME=$(APPNAME)_linux_amd64
	endif
	LDFLAGs=-ldflags '-X "github.com/kabacloud/cloudnativehomework4-module10/environment.BuildInfo='"$(BUILD_INFO)"'" '
endif

# 容器信息
ALPINE_TAG=3.14
IMAGENAME=$(APPNAME)
IMAGETAG=$(APPVERSION)
CONTAINERNAME=$(IMAGENAME)_$(IMAGETAG)
CONTAINERNAME=$(IMAGENAME)_$(IMAGETAG)
DOCKERREPO=kabacloud

# 声明命令列表，避免和同名文件冲突
.PHONY: all web clean format mod build ut run dev buildimage dockerrm dockerrun help

all: help

clean: # 清理构筑环境
	$(GOCLEAN)
format: # 格式化代码
	$(GOFMT) .
mod: ## 整理vendor依赖包
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor
build: clean format ## 编译应用
	$(GOBUILD) -o $(BINARY_NAME) -v $(LDFLAGs)
ut: build ## 单元测试并生成测试报告
	$(GOTEST) -v -test.run TestUnit_ -coverprofile=coverUnit.out ./...
	$(GOTOOL) cover -html=coverUnit.out -o coverUnit.html
	rm coverUnit.out
run: build # 启动主命令
	./$(BINARY_NAME)
serve: build ## 本机启动服务
	HWENV=Develop VERSION=1.0.0 ./$(BINARY_NAME) serve

dockerbase: ## 编译基础镜像
	@echo "==== build image alpine with curl ===="
	docker build --build-arg ALPINE_TAG="$(ALPINE_TAG)" -t alpine:$(ALPINE_TAG)-curl -f Base-curl.Dockerfile .
	docker tag alpine:$(ALPINE_TAG)-curl kabacloud/alpine:$(ALPINE_TAG)-curl
	docker push kabacloud/alpine:$(ALPINE_TAG)-curl
	@echo "==== build image alpine with sshpass ===="
	docker build --build-arg ALPINE_TAG="$(ALPINE_TAG)" -t alpine:$(ALPINE_TAG)-rsync-sshpass -f Base-rsync-sshpass.Dockerfile .
	docker tag alpine:$(ALPINE_TAG)-rsync-sshpass kabacloud/alpine:$(ALPINE_TAG)-rsync-sshpass
	docker push kabacloud/alpine:$(ALPINE_TAG)-rsync-sshpass
docker-lint: # 检查dockerfile
	hadolint --config .hadolint.yaml Dockerfile
docker-build: docker-lint # 应用打包成镜像
	@echo "==== 编译镜像 ===="
	docker build --build-arg BUILD_INFO="$(BUILD_INFO)" --build-arg BASE_IMAGE="kabacloud/alpine:$(ALPINE_TAG)-curl" -t $(IMAGENAME):$(IMAGETAG) . 
docker-push: docker-build ## 推送应用镜像
	@echo "==== 推送镜像 ===="
	docker tag $(IMAGENAME):$(IMAGETAG) kabacloud/$(IMAGENAME):$(IMAGETAG)
	docker push kabacloud/$(IMAGENAME):$(IMAGETAG)
docker-clean: # 删除容器
	@echo "==== 清除容器 ===="
	docker rm $(CONTAINERNAME)
docker-run: docker-clean ## 执行应用容器
	@echo "==== 启动容器 ===="
	docker run -t -p 8000:8000 --name $(CONTAINERNAME) kabacloud/$(IMAGENAME):$(IMAGETAG)
docker-start: # 启动容器
	docker container start $(CONTAINERNAME)
docker-stop: # 停止容器
	docker container stop $(CONTAINERNAME)

# okteto部署前要确认kubeconfig是否有效，无效的话执行如下操作：
# $ okteto login
#  ✓  Logged in as woshizilong
# $ okteto namespace
#  ✓  Updated context 'cloud_okteto_com' in '/Users/songzilong/.kube/config'
# 至此名为cloud_okteto_com的配置追加到.kube/config中了，可以针对okteto执行kubectl操作了
okteto-build: # 编译okteto部署用镜像
	GOROOT=$(GOROOT) KO_DOCKER_REPO=$(DOCKERREPO) KO_CONFIG_PATH=.ko.yaml GOFLAGS="-ldflags=-X=github.com/kabacloud/cloudnativehomework4-module10/environment.BuildInfo=$(BUILD_INFO)" ko resolve -f k8s-okteto-template.yaml > k8s-okteto.yaml
okteto-lint: okteto-build # 检查生成的k8s文件是否符合k8s规范
	kube-linter lint k8s-okteto.yaml --config .kube-linter-okteto.yaml
#`kubectl delete -f k8s-okteto.yaml`
okteto-apply: okteto-lint ## 部署到 okteto（k8s环境）
	kubectl apply -f k8s-okteto.yaml
	@echo "You can access https://httpserver-homework-woshizilong.cloud.okteto.net/info"
okteto-debug: # https://okteto.com/docs/samples/golang/
	okteto debug --port=8000 --image=kabacloud/okteto-debug:1.0.0

k8s-build: # 编译homework部署用镜像
	GOROOT=$(GOROOT) KO_DOCKER_REPO=$(DOCKERREPO) KO_CONFIG_PATH=.ko.yaml GOFLAGS="-ldflags=-X=github.com/kabacloud/cloudnativehomework4-module10/environment.BuildInfo=$(BUILD_INFO)" ko resolve -f k8s-homework-template.yaml > k8s-homework.yaml
k8s-lint: k8s-build # 检查生成的k8s文件是否符合k8s规范
	kube-linter lint k8s-homework.yaml --config .kube-linter-homework.yaml
#`kubectl delete -f k8s-homework.yaml`
k8s-apply: k8s-lint ## 部署到k8s环境
	kubectl apply -f k8s-homework.yaml
	@echo "You can access https://111/info"

# go install github.com/grafana/dashboard-linter
lint-dashboard: # 检查Grafana的Dashboard设定
	dashboard-linter lint dashboard.json

help: ## 帮助信息
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-16s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)
