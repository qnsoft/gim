# 二进制文件编译
FROM golang:stretch as base

MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOPROXY="https://goproxy.io"

WORKDIR /go/gim
COPY ./src/client/client.go .
RUN go build -a -installsuffix cgo -o app client.go

# 编译体积更小的分发版本
FROM scratch
MAINTAINER WangXiaoQing "https://github.com/wangxiaoqiange"

COPY --from=base /go/gim/app .

ENTRYPOINT ["/app"]
CMD ["--help"]
