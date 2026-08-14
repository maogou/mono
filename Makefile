APP_NAME := go_template
GO_MODULE := go_template/cmd/go_template
VERSION := $(shell git rev-parse --short HEAD)
DOCKER_IMAGE := go_template
HOST_PORT := 8073
CONTAINER_PORT := 8073
CLEAN_IMAGE = $(shell docker images -a | grep ${APP_NAME} | awk '{print $$3}')
CLEAN_CONTAINER = $(shell docker ps -a | grep ${APP_NAME} | awk '{print $$1}')


.PHONY: build run docker-build docker-run clean lint

build:
	@echo "编译二进制文件..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ${APP_NAME} ${GO_MODULE}

run:
	@echo "运行服务..."
	go run ./cmd/${APP_NAME}/ start

docker-build:
	@echo "构建Docker镜像..."
	docker buildx build \
    	--platform linux/arm64 \
		--build-arg APP_NAME=${APP_NAME}_${VERSION} \
		-t ${DOCKER_IMAGE}:${VERSION} \
		.

docker-run:
	@echo "开始运行容器..."
	docker run -d \
		--name ${APP_NAME}_${VERSION} \
		-p ${HOST_PORT}:${CONTAINER_PORT} \
		${DOCKER_IMAGE}:${VERSION}

clean:
	@echo "清理构建产物..."
	rm -f ${APP_NAME}
	@echo "清理容器..."
	docker rm -f ${CLEAN_CONTAINER} || true
	@echo "清理Docker镜像..."
	docker rmi -f ${CLEAN_IMAGE} || true
	@echo "清理Docker镜像缓存..."
	docker image prune -f

lint:
	@echo "代码检查..."
	golangci-lint run  ./...

MOCKGEN := go run go.uber.org/mock/mockgen@v0.6.0

# 覆盖率报告用浏览器打开,按操作系统选择命令
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
OPEN_CMD := open
endif
ifeq ($(UNAME_S),Linux)
OPEN_CMD := xdg-open
endif

.PHONY: mock test coverage

mock:
	@echo "生成mock文件..."
	$(MOCKGEN) -source=internal/repository/repository.go -destination=test/mocks/repository/repository.go
	$(MOCKGEN) -source=internal/repository/demo_repo.go -destination=test/mocks/repository/demo_repo.go
	$(MOCKGEN) -source=internal/service/demo_service.go -destination=test/mocks/service/demo_service.go

test:
	@echo "运行测试..."
	go test -coverpkg=./internal/repository,./internal/service -coverprofile=./coverage.out ./test/server/...

coverage: test
	@echo "生成覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@if [ -n "$(OPEN_CMD)" ]; then echo "打开 coverage.html ..."; $(OPEN_CMD) coverage.html; else echo "请手动打开 coverage.html"; fi
