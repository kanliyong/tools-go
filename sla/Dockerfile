FROM golang:1.15.5

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /go/src/tools
ADD main.go .
RUN go mod init
RUN go build -o sla main.go

FROM ubuntu:18.04

RUN apt update && apt-get install -y ca-certificates

COPY --from=0 /go/src/tools/sla /sla
ENTRYPOINT ["/sla"]