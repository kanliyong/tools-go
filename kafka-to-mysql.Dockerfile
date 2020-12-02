FROM golang:1.15.5
WORKDIR /go/src/tools
ENV GOPROXY="https://goproxy.cn,direct"

ADD kafka-to-mysql.go .
ADD go.mod .

RUN go build -o kafka-to-mysql kafka-to-mysql.go

FROM ubuntu:18.04
COPY --from=0 /go/src/tools/kafka-to-mysql /kafka-to-mysql
ENTRYPOINT ["/kafka-to-mysql"]