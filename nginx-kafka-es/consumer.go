// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type Consumer struct {
	Brokers         []string
	TopicName       string
	GroupID         string
	MessageCallback func(message *kafka.Message) ([]byte, error)

	Indexer esutil.BulkIndexer
	reader  *kafka.Reader

	startTime     time.Time
	totalMessages int64
	totalErrors   int64
	totalBytes    int64
}

func (c *Consumer) Run(ctx context.Context) (err error) {
	if c.Indexer == nil {
		panic(fmt.Sprintf("%T.Indexer is nil", c))
	}
	c.startTime = time.Now()

	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Brokers,
		GroupID: c.GroupID,
		Topic:   c.TopicName,
		// MinBytes: 1e+6, // 1MB
		// MaxBytes: 5e+6, // 5MB

		ReadLagInterval: 1 * time.Second,
	})

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("reader: %s", err)
		}
		// log.Printf("%v/%v/%v:%s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		value, err := c.MessageCallback(&msg)

		item := esutil.BulkIndexerItem{
			Action: "create",
			Body:   bytes.NewReader(value),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				// log.Printf("Indexed %s/%s", res.Index, res.DocumentID)
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("failure %s", err)
				} else {
					if res.Error.Type != "" {
						log.Printf("%s:%s", res.Error.Type, res.Error.Reason)
					} else {
						log.Printf("%s/%s %s (%d)", res.Index, res.DocumentID, res.Result, res.Status)
					}

				}
			},
		}
		if err := c.Indexer.Add(ctx, item); err != nil {
			log.Printf("indexer: %s", err)
		}
	}
	c.reader.Close()
	c.Indexer.Close(ctx)

	return nil
}

type Stats struct {
	Duration      time.Duration
	TotalLag      int64
	TotalMessages int64
	TotalErrors   int64
	TotalBytes    int64
	Throughput    float64
}

func (c *Consumer) Stats() Stats {
	if c.reader == nil || c.Indexer == nil {
		return Stats{}
	}

	duration := time.Since(c.startTime)
	readerStats := c.reader.Stats()

	c.totalMessages += readerStats.Messages
	c.totalErrors += readerStats.Errors
	c.totalBytes += readerStats.Bytes

	rate := float64(c.totalMessages) / duration.Seconds()

	return Stats{
		Duration:      duration,
		TotalLag:      readerStats.Lag,
		TotalMessages: c.totalMessages,
		TotalErrors:   c.totalErrors,
		TotalBytes:    c.totalBytes,
		Throughput:    rate,
	}
}
