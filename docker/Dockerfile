FROM golang:latest as base

ENV GO111MODULE=on GOPROXY="https://goproxy.io"

WORKDIR /go/gim

COPY build .

RUN go build -o Gim src/main.go

FROM alpine:latest

WORKDIR /

COPY --from=base /go/gim/Gim .

RUN chmod +x Gim

ENTRYPOINT ["/Gim"]
