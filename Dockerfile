# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装ca-certificates（HTTPS需要）
RUN apk --no-cache add ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/gateway .

# 复制配置文件
COPY config/config.example.yaml ./config/config.yaml

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 18168

# 运行
CMD ["./gateway", "-config", "./config/config.yaml"]
