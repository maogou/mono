FROM golang:1.26-alpine AS builder
ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update --no-cache && apk add --no-cache tzdata
WORKDIR /build
COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0  go build  -ldflags="-s -w" -o go_template ./cmd/template
FROM alpine:3.17 AS final
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /build/config/go_template.yaml /app/config/go_template.yaml
COPY --from=builder /build/go_template /app/go_template
CMD ["./go_template" ,"start"]
