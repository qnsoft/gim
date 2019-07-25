# 多阶段构建, 基础镜像用来编译源代码
FROM golang:stretch as base

MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOPROXY="https://goproxy.io"

WORKDIR /go/gim
COPY . .
RUN go build -a -installsuffix cgo -o app src/main.go

# release
FROM alpine:latest
MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

RUN mkdir -p /code/src

WORKDIR /code
COPY --from=base /go/gim/app .
COPY --from=base /go/gim/src/config.json ./src

# 存放 config.json
VOLUME /src

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

CMD ["./app"]
