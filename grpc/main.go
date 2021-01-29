package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"grpc/dubbo"

	"google.golang.org/grpc"
)

func main() {

	id := os.Args[1]
	conn, err := grpc.Dial("localhost:5050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := dubbo.NewProxyServiceClient(conn)
	value := fmt.Sprintf("[%s]", id)
	request := dubbo.ProxyRequest{
		InterfaceName: "eqxiu.mall.product.service.api.ProductServiceApi",
		Version:       "1.1.0",
		Method:        "getProductDetailBySourceId",
		Types:         []string{"int"},
		Values:        value,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reply, err := c.InvokeProxy(ctx, &request)

	if err == nil {
		result := make(map[string]interface{})
		json.Unmarshal([]byte(reply.GetResult()), &result)
		println(result)
	} else {
		log.Println(err)
	}
}
