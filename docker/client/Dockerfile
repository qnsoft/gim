# 多阶段构建, 基础镜像用来编译源代码
FROM golang:stretch as base

MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOPROXY="https://goproxy.io"

WORKDIR /go/gim
COPY ./src/client/client.go .
RUN go build -a -installsuffix cgo -o app client.go

# release
FROM scratch
MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

COPY --from=base /go/gim/app .

ENTRYPOINT ["/app"]
CMD ["--help"]
