
## linux build
```bash
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o output main.go
```
## mac build

```bash
env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o output main.go
```