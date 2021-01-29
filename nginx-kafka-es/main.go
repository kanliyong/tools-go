package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

var (
	servers      string
	group_id     string
	topic        string
	numWorker    int
	indexerError error
	wg           sync.WaitGroup
)

func init() {

	flag.StringVar(&servers,
		"servers",
		"hadoop104.eqxiu.com:9092",
		"kafka bootstrap servers")
	flag.StringVar(&group_id, "group_id", "gateway_original_mysql", "kafka consumer group id")
	flag.StringVar(&topic, "topic", "gateway_original", "kafka consumer topic")
	flag.IntVar(&numWorker, "numWorker", 1, "number of worker")
	flag.Parse()

	log.Printf("group_id = %s", group_id)
	log.Printf("topic = %s", topic)
	log.Printf("servers = %s", servers)
	log.Printf("numCPU %d", runtime.NumCPU())
}

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

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
	// numWorkers := 0
	flushBytes := 0
	var indexers []esutil.BulkIndexer
	var consumers []*Consumer

	for i := 0; i < numWorker; i++ {
		idx, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Index:      url.QueryEscape("<waf_nginx_log_{now/d}>"),
			Client:     es,
			NumWorkers: runtime.NumCPU(),
			FlushBytes: int(flushBytes),
			// Elastic APM: Instrument the flush operations and capture errors
			OnFlushStart: func(ctx context.Context) context.Context {
				log.Printf("start flushing")
				return ctx
			},
			OnFlushEnd: func(ctx context.Context) {
				log.Printf("end flushing")
			},
			OnError: func(ctx context.Context, err error) {
				indexerError = err
				log.Printf("flushing error %s", err)
			},
		})
		if err != nil {
			log.Fatalf("ERROR: NewBulkIndexer(): %s", err)
		}
		indexers = append(indexers, idx)
	}

	for i := 0; i < numWorker; i++ {
		consumers = append(consumers,
			&Consumer{
				Brokers: []string{
					"hadoop104.eqxiu.com:9092",
					"hadoop105.eqxiu.com:9092",
					"hadoop106.eqxiu.com:9092",
				},
				TopicName:       topic,
				GroupID:         group_id,
				MessageCallback: onMessage,
				Indexer:         indexers[i]})
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

	reporter := time.NewTicker(5000 * time.Millisecond)
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

func report(consumers []*Consumer, indexers []esutil.BulkIndexer) string {
	var b strings.Builder

	for i, c := range consumers {
		s := c.Stats()
		log := fmt.Sprintf("Consumer %d lagging=%s msg/sec=%s received=%s bytes=%s errors=%s",
			i+1,
			humanize.Comma(s.TotalLag),
			humanize.FtoaWithDigits(s.Throughput, 2),
			humanize.Comma(s.TotalMessages),
			humanize.Bytes(uint64(s.TotalBytes)),
			humanize.Comma(s.TotalErrors),
		)
		fmt.Fprint(&b, log)
		fmt.Fprint(&b, "\r\n")
	}

	for i, x := range indexers {
		s := x.Stats()
		log := fmt.Sprintf("Indexer %d added=%s flushed=%s failed=%s",
			i+1,
			humanize.Comma(int64(s.NumAdded)),
			humanize.Comma(int64(s.NumFailed)),
			humanize.Comma(int64(s.NumFailed)),
		)
		fmt.Fprintf(&b, log)
		fmt.Fprint(&b, "\r\n")
	}
	return b.String()
}
