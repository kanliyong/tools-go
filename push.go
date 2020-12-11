package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/push"
	"net/http"
	"time"
)

var (
	completionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_last_completion_timestamp_seconds",
		Help: "The timestamp of the last completion of a DB backup, successful or not.",
	})
	successTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_last_success_timestamp_seconds",
		Help: "The timestamp of the last successful completion of a DB backup.",
	})
	duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_duration_seconds",
		Help: "The duration of the last DB backup in seconds.",
	})
	records = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_records_processed",
		Help: "The number of records processed in the last DB backup.",
	})
)


func main() {
	pushGateWay := "http://localhost:9091"
	job := "job"

	registry := prometheus.NewRegistry()
	registry.MustRegister(completionTime, duration, records)
	pusher := push.New(pushGateWay, job).Gatherer(registry)

	ticker := time.NewTicker(15 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:

				duration.Set(10)
				completionTime.SetToCurrentTime()
				// pusher.Add()
				if err := pusher.Add(); err != nil {
					fmt.Println("Could not push completion time to Pushgateway:", err)
				}else{
					fmt.Println("push completion to Pushgateway:")
				}
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	err := http.ListenAndServe(":9123", nil)
	if err != nil {
		panic(err)
	}
}
