# 二进制文件编译
FROM golang:stretch as base

MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOPROXY="https://goproxy.io"

WORKDIR /go/gim
COPY . .
RUN go build -a -installsuffix cgo -o app src/main.go

# 编译体积更小的分发版本
FROM alpine:latest
MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

RUN mkdir -p /code/src

WORKDIR /code
COPY --from=base /go/gim/app .
COPY --from=base /go/gim/src/config.json ./src

# Store config.json
VOLUME /src

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

CMD ["./app"]
