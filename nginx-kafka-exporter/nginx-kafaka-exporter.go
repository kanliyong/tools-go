package main

import (
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"strconv"
	"strings"
	"time"
)

/**
docker build -t ccr.ccs.tencentyun.com/eqxiu/nginx-kafaka-exporter -f nginx-kafaka-exporter.Dockerfile .
docker push ccr.ccs.tencentyun.com/eqxiu/nginx-kafaka-exporter



./nginx-kafaka-exporter -servers=hadoop104.eqxiu.com:9092,hadoop105.eqxiu.com:9092,hadoop106.eqxiu.com:9092 \
-group_id=gateway_original_test\
-topic=gateway_original

*/
var (
	waf_http_response_count_total = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "waf_http_response_count_total",
		Help: "Amount of processed HTTP requests",
	},[]string{"domain","method","status"})

	waf_http_response_size_bytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "waf_http_response_size_bytes_total",
		Help: "Total amount of transferred bytes",
	},[]string{"domain","method","status"})

	waf_http_response_time_seconds_hist = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "waf_http_response_time_seconds_hist",
		Help: "Time needed by NGINX to handle requests",
	},[]string{"domain","method","status"})

	waf_http_response_time_seconds = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "waf_http_response_time_seconds",
		Help: "Time needed by NGINX to handle requests",
	},[]string{"domain","method","status"})
)

func parseTime(input string) time.Time {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", input)
	if err != nil {
		return time.Now()
	}
	return t
}

var (
	servers string
	group_id string
	topic string
)

func init(){

	flag.StringVar(&servers,
		"servers",
		"hadoop104.eqxiu.com:9092,hadoop105.eqxiu.com:9092,hadoop106.eqxiu.com:9092",
		"kafka bootstrap servers")
	flag.StringVar(&group_id, "group_id","group_id_test", "kafka consumer group id")
	flag.StringVar(&topic, "topic","topic_name", "kafka consumer topic")
	flag.Parse()

}
func insertBatch(log NginxLog) {
	domain := log.Host
	method := log.Method
	status := log.Code

	waf_http_response_count_total.WithLabelValues(domain, method, strconv.Itoa(status)).Inc()
	waf_http_response_size_bytes.WithLabelValues(domain, method, strconv.Itoa(status)).Add(log.ResponseBodySize)
	waf_http_response_time_seconds_hist.WithLabelValues(domain, method, strconv.Itoa(status)).Observe(log.RequestTime)
	waf_http_response_time_seconds.WithLabelValues(domain, method, strconv.Itoa(status)).Observe(log.RequestTime)
}



type NginxLog struct {
	RemoteIp         string
	JessionId        string
	Host             string
	Method           string
	Uri              string
	Code             int
	RequestTime      float64
	ResponseBodySize float64
	Time             time.Time
	Agent            string
	Referer          string
	UpstreamAddr     string
	UpstreamStatus   int
	Tracker          string
}

func onMessage(e *kafka.Message) NginxLog {

	s := string(e.Value)
	rs := strings.Split(s, "#|#")

	code, _ := strconv.Atoi(rs[8])
	requestTime, _ := strconv.ParseFloat(rs[11], 8)
	ResponseBodySize, _ := strconv.ParseFloat(rs[15], 8)

	data := NginxLog{
		RemoteIp:         rs[0],
		JessionId:        rs[2],
		Host:             rs[3],
		Method:           rs[4],
		Uri:              maxLength(rs[5], 250),
		Code:             code,
		RequestTime:      requestTime,
		ResponseBodySize: ResponseBodySize,
		Time:             parseTime(rs[1]),
		Agent:            maxLength(rs[9], 250),
		Referer:          maxLength(rs[10], 250),
		UpstreamAddr:     rs[12],
		UpstreamStatus:   1,
		Tracker:          rs[18],
	}
	return data
}

func maxLength(content string, maxLen int) string {
	asRunes := []rune(content)
	if len(asRunes) > maxLen {
		return string(asRunes[:maxLen])
	} else {
		return content
	}
}

func consumer() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": servers,
		"group.id":          group_id,
		"auto.offset.reset": "smallest"})

	if err != nil {
		panic(err)
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil{
		panic(err)
	}
	defer consumer.Close()

	for {
		ev := consumer.Poll(0)
		switch e := ev.(type) {
		case *kafka.Message:
			data := onMessage(e)
			insertBatch(data)
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			break
		default:
			//fmt.Printf("Ignored %v\n", e)
		}
	}


}



func main() {

	go consumer()

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9123", nil)
	if err != nil {
		panic(err)
	}
}
