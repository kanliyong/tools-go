package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

/**
docker build -t ccr.ccs.tencentyun.com/eqxiu/nginx-kafaka-exporter -f Dockerfile .
docker push ccr.ccs.tencentyun.com/eqxiu/nginx-kafaka-exporter



./nginx-kafaka-exporter -servers=hadoop104.eqxiu.com:9092,hadoop105.eqxiu.com:9092,hadoop106.eqxiu.com:9092 \
-group_id=gateway_original_test\
-topic=gateway_original

*/

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
	indexerError error
	wg  sync.WaitGroup
)

func init(){

	flag.StringVar(&servers,
		"servers",
		"hadoop104.eqxiu.com:9092",
		"kafka bootstrap servers")
	flag.StringVar(&group_id, "group_id","gateway_original_mysql", "kafka consumer group id")
	flag.StringVar(&topic, "topic","gateway_original", "kafka consumer topic")
	flag.Parse()

	log.Printf("group_id = %s", group_id)
	log.Printf("topic = %s", topic)
	log.Printf("servers = %s", servers)
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

func onMessage(e *kafka.Message) ([]byte, error) {

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
		Uri:              rs[5],
		Code:             code,
		RequestTime:      requestTime,
		ResponseBodySize: ResponseBodySize,
		Time:             parseTime(rs[1]),
		Agent:            rs[9],
		Referer:          rs[10],
		UpstreamAddr:     rs[12],
		UpstreamStatus:   1,
		Tracker:          rs[18],
	}
	return json.Marshal(data)
}

func maxLength(content string, maxLen int) string {
	asRunes := []rune(content)
	if len(asRunes) > maxLen {
		return string(asRunes[:maxLen])
	} else {
		return content
	}
}

func main() {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://10.0.10.34:9200",
			"http://10.0.10.35:9200",
			"http://10.0.10.43:9200",
			"http://10.0.20.30:9200",
			"http://10.0.20.31:9200",
		},
		RetryOnStatus: []int{502, 503, 504, 429}, // Add 429 to the list of retryable statuses
		RetryBackoff:  func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
		MaxRetries:    5,
		EnableMetrics: true,
	})
	if err != nil {
		log.Fatalf("Error: NewClient(): %s", err)
	}
	numIndexers := 5
	numConsumers := 10
	numWorkers := 0
	flushBytes := 0
	var indexers  []esutil.BulkIndexer
	var consumers []*Consumer

	for i := 1; i <= numIndexers; i++ {
		idx, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Index:      fmt.Sprintf("waf_nginx_log_%s", time.Now().Format("2006-01-02")),
			Client:     es,
			NumWorkers: numWorkers,
			FlushBytes: int(flushBytes),
			// Elastic APM: Instrument the flush operations and capture errors
			OnFlushStart: func(ctx context.Context) context.Context {
				txn := apm.DefaultTracer.StartTransaction("Bulk", "indexing")
				return apm.ContextWithTransaction(ctx, txn)
			},
			OnFlushEnd: func(ctx context.Context) {
				apm.TransactionFromContext(ctx).End()
			},
			OnError: func(ctx context.Context, err error) {
				indexerError = err
				apm.CaptureError(ctx, err).Send()
			},
		})
		if err != nil {
			log.Fatalf("ERROR: NewBulkIndexer(): %s", err)
		}
		indexers = append(indexers, idx)
	}

	for i := 1; i <= numConsumers; i++ {
		consumers = append(consumers,
			&Consumer{
				Brokers: []string{
					"hadoop104.eqxiu.com:9092",
					"hadoop105.eqxiu.com:9092",
					"hadoop106.eqxiu.com:9092",
				},
				TopicName: topic,
				GroupID: group_id,
				MessageCallback: onMessage,
				Indexer:   indexers[i%numIndexers]})
	}

	for _, c := range consumers {
		wg.Add(1)
		go func(c *Consumer) {
			defer wg.Done()
			if err := c.Run(context.TODO()); err != nil {
				log.Fatalf("ERROR: Consumer: %s", err)
			}
		}(c)
	}

	reporter := time.NewTicker(500 * time.Millisecond)
	defer reporter.Stop()
	go func() {
		for {
			select {
			case <-reporter.C:
				fmt.Print(report(consumers, indexers))
			}
		}
	}()
	wg.Add(1)
	wg.Wait()

}


func report(
	consumers []*Consumer,
	indexers []esutil.BulkIndexer,
) string {
	var (
		b strings.Builder

		value    string
		currRow  = 1
		numCols  = 6
		colWidth = 20

		divider = func(last bool) {
			fmt.Fprintf(&b, "\033[%d;0H", currRow)
			fmt.Fprint(&b, "┣")
			for i := 1; i <= numCols; i++ {
				fmt.Fprint(&b, strings.Repeat("━", colWidth))
				if last && i == 5 {
					fmt.Fprint(&b, "┷")
					continue
				}
				if i < numCols {
					fmt.Fprint(&b, "┿")
				}
			}
			fmt.Fprint(&b, "┫")
			currRow++
		}
	)

	fmt.Print("\033[2J\033[K")
	fmt.Printf("\033[%d;0H", currRow)

	fmt.Fprint(&b, "┏")
	for i := 1; i <= numCols; i++ {
		fmt.Fprint(&b, strings.Repeat("━", colWidth))
		if i < numCols {
			fmt.Fprint(&b, "┯")
		}
	}
	fmt.Fprint(&b, "┓")
	currRow++


	for i, c := range consumers {
		fmt.Fprintf(&b, "\033[%d;0H", currRow)
		value = fmt.Sprintf("Consumer %d", i+1)
		fmt.Fprintf(&b, "┃ %-*s│", colWidth-1, value)
		s := c.Stats()
		value = fmt.Sprintf("lagging=%s", humanize.Comma(s.TotalLag))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("msg/sec=%s", humanize.FtoaWithDigits(s.Throughput, 2))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("received=%s", humanize.Comma(s.TotalMessages))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("bytes=%s", humanize.Bytes(uint64(s.TotalBytes)))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("errors=%s", humanize.Comma(s.TotalErrors))
		fmt.Fprintf(&b, " %-*s┃", colWidth-1, value)
		currRow++
		divider(i == len(consumers)-1)
	}

	for i, x := range indexers {
		fmt.Fprintf(&b, "\033[%d;0H", currRow)
		value = fmt.Sprintf("Indexer %d", i+1)
		fmt.Fprintf(&b, "┃ %-*s│", colWidth-1, value)
		s := x.Stats()
		value = fmt.Sprintf("added=%s", humanize.Comma(int64(s.NumAdded)))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("flushed=%s", humanize.Comma(int64(s.NumFlushed)))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		value = fmt.Sprintf("failed=%s", humanize.Comma(int64(s.NumFailed)))
		fmt.Fprintf(&b, " %-*s│", colWidth-1, value)
		if indexerError != nil {
			value = "err=" + indexerError.Error()
			if len(value) > 2*colWidth {
				value = value[:2*colWidth]
			}
		} else {
			value = ""
		}
		fmt.Fprintf(&b, " %-*s┃", 2*colWidth, value)
		currRow++
		if i < len(indexers)-1 {
			divider(true)
		}
	}

	fmt.Fprintf(&b, "\033[%d;0H", currRow)
	fmt.Fprint(&b, "┗")
	for i := 1; i <= numCols; i++ {
		fmt.Fprint(&b, strings.Repeat("━", colWidth))
		if i == 5 {
			fmt.Fprint(&b, "━")
			continue
		}
		if i < numCols {
			fmt.Fprint(&b, "┷")
		}
	}
	fmt.Fprint(&b, "┛")
	currRow++

	return b.String()
}