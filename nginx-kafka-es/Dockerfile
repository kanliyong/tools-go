FROM golang:1.15.5

ENV GOPROXY="https://goproxy.cn,direct"
WORKDIR /go/src/tools

ADD consumer.go .
ADD main.go .
RUN go mod init
RUN go build -o main main.go consumer.go

FROM ubuntu:18.04
COPY --from=0 /go/src/tools/main /main
ENTRYPOINT ["/main"]