package dubbo

import (
	context "context"
	_ "encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	grpc "google.golang.org/grpc"
)

/*
go test -bench=. ./dubbo
*/
func conn(url string) ProxyServiceClient {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := NewProxyServiceClient(conn)
	return c
}
func get(c ProxyServiceClient) {

	value := fmt.Sprintf("[%s]", "5080")
	request := ProxyRequest{
		InterfaceName: "eqxiu.mall.product.service.api.ProductServiceApi",
		Version:       "1.1.0",
		Method:        "getProductDetailBySourceId",
		Types:         []string{"int"},
		Values:        value,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.InvokeProxy(ctx, &request)

	if err == nil {
		// println(string(reply.GetResult()))
	} else {
		log.Println(err)
	}
}
func BenchmarkDirect(b *testing.B) {
	c := conn("10.0.6.41:5050")
	for i := 0; i < b.N; i++ {
		get(c)
	}
}

func BenchmarkNginx(b *testing.B) {
	c := conn("grpc-dubbo.internal.eqxiu.com:5050")
	for i := 0; i < b.N; i++ {
		get(c)
	}
}

func BenchmarkNginxDirect(b *testing.B) {
	c := conn("10.0.20.140:5151")
	for i := 0; i < b.N; i++ {
		get(c)
	}
}
