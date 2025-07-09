# 第一阶段：构建阶段
FROM golang:1.23 AS builder

WORKDIR /app

# 使用国内模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 将 go.mod 和 go.sum 复制进去（如果有的话），以利用缓存机制
COPY go.mod ./
# 如果有 go.sum，也复制进来
# COPY go.sum ./
RUN go mod download

# 将所有代码复制到容器中
COPY . .

# 编译你的项目，假设入口文件为 main.go，你可以根据实际情况调整
RUN CGO_ENABLED=0 GOOS=linux go build -o siwuai ./cmd/server/main.go

# 第二阶段：生成最小镜像
FROM alpine:latest
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/siwuai .

# 如有需要，也可以复制配置文件到镜像内（或者选择在运行时通过挂载来提供）
# COPY config/ /app/config/

# 开放端口（这里配置文件中定义的服务端口为50051）
EXPOSE 50051

# 启动应用
CMD ["./siwuai"]
