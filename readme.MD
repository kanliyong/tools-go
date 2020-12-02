
## linux build
```bash
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o output main.go
```
## mac build

```bash
env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o output main.go
go build -gcflags=all="-N -l" main.go
```

## 依赖

### mysql 
https://github.com/go-sql-driver/mysql

### kafka
https://docs.confluent.io/clients-confluent-kafka-go/current/index.html

