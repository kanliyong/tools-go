FROM golang:1.15.5

ENV GOPROXY="https://goproxy.cn,direct"
WORKDIR /go/src/tools

ADD consumer.go .
ADD main.go .
ADD nginx.go .
RUN go mod init
RUN go build -o main main.go consumer.go nginx.go

FROM ubuntu:18.04

## 时区设置
ENV TZ=Asia/Shanghai
RUN apt update \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
    && apt-get -y install tzdata \
    && apt-get -y clean \
    && apt-get -y autoclean 
    
COPY --from=0 /go/src/tools/main /main
ENTRYPOINT ["/main"]