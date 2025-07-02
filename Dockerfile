# 构建前端
FROM node AS web_image

# 华为源
RUN npm config set registry https://repo.huaweicloud.com/repository/npm/

RUN npm install pnpm -g

WORKDIR /build

COPY ./package.json /build
COPY ./pnpm-lock.yaml /build

RUN pnpm install

COPY . /build

RUN pnpm run build

# 构建hello_favicon服务
FROM golang:1.23.7-alpine3.20 AS hello_favicon_builder

WORKDIR /app

# 将hello_favicon项目代码复制到容器内的工作目录
COPY hello_favicon/ .

# 安装CA证书、构建二进制文件
RUN apk add --no-cache ca-certificates upx && \
    go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/hello_favicon_main . && \
    upx --best --lzma /app/hello_favicon_main

# 构建sun-panel服务
FROM golang:1.21-alpine3.18 as sun_panel_builder

WORKDIR /build

COPY ./service .

RUN apk add --no-cache bash curl gcc git musl-dev

RUN go env -w GO111MODULE=on && \
    export PATH=$PATH:/go/bin && \
    go install -a -v github.com/go-bindata/go-bindata/...@latest && \
    go install -a -v github.com/elazarl/go-bindata-assetfs/...@latest && \
    go-bindata-assetfs -o=assets/bindata.go -pkg=assets assets/... && \
    go build -o sun-panel --ldflags="-X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" main.go

# 最终运行镜像
FROM alpine:latest

WORKDIR /app

# 安装supervisord和其他必要工具
RUN apk add --no-cache bash ca-certificates su-exec tzdata supervisor && \
    mkdir -p /var/log/supervisor

# 复制前端构建结果
COPY --from=web_image /build/dist /app/web

# 复制sun-panel服务
COPY --from=sun_panel_builder /build/sun-panel /app/sun-panel
RUN chmod +x /app/sun-panel

# 复制hello_favicon服务和相关文件
COPY --from=hello_favicon_builder /app/hello_favicon_main /app/hello_favicon_main
COPY --from=hello_favicon_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY hello_favicon/static /app/hello_favicon/static
COPY hello_favicon/templates /app/hello_favicon/templates
RUN chmod +x /app/hello_favicon_main

# 复制supervisord配置文件
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# 暴露端口
EXPOSE 3000 3002

# 初始化sun-panel配置
RUN /app/sun-panel -config

# 启动supervisord
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]