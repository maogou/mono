APP_NAME := go_template
GO_MODULE := go_template/cmd/go_template
VERSION := $(shell git rev-parse --short HEAD)
DOCKER_IMAGE := go_template
HOST_PORT := 8073
CONTAINER_PORT := 8073
CLEAN_IMAGE := $(shell docker images -a | grep ${APP_NAME} | awk '{print $$3}')
CLEAN_CONTAINER := $(shell docker ps -a | grep ${APP_NAME} | awk '{print $$1}')


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
