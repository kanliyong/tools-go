FROM golang:1.15.5

ENV GOPROXY="https://goproxy.cn,direct"
WORKDIR /go/src/tools

ADD nginx-kafaka-exporter.go .
RUN go mod init
RUN go build -o nginx-kafaka-exporter nginx-kafaka-exporter.go

FROM ubuntu:18.04
COPY --from=0 /go/src/tools/nginx-kafaka-exporter /nginx-kafaka-exporter
ENTRYPOINT ["/nginx-kafaka-exporter"]