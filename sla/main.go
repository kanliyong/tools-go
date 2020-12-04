package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GraphqlParam struct {
	Query string	`json:"query"`
	Variables		`json:"variables"`
}

type Variables struct {
	Duration	`json:"duration"`
	Condition	`json:"condition"`
}

type Duration struct {
	Start string	`json:"start"`
	End string		`json:"end"`
	Step string		`json:"step"`
}
type Condition struct {
	Name string		`json:"name"`
	Normal bool		`json:"normal"`
	Order string	`json:"order"`
	ParentService string	`json:"parentService"`
	Scope string	`json:"scope"`
	TopN string		`json:"topN"`
}

type Result struct {
	Data SortMetrics `json:"data"`
}

type SortMetrics struct {
	Metrics  []SLA `json:"sortMetrics"`
}

type SLA struct {
	ID string		`json:"id"`
	Name string		`json:"name"`
	Value string	`json:"value"`
}

var dubbo_service_sla = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "dubbo_service_sla",
	Help: "dubbo_service_sla",
},[]string{"name"})

func main() {

	go fetch()
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9123", nil)
	if err != nil {
		panic(err)
	}
	log.Printf("start on port %d", 9123)
}

func fetch(){


	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:

				start := time.Now().Add(-15 * time.Minute)
				end := time.Now()
				layout := "2006-01-02 1504"

				resultData := getSlaData(start.Format(layout),end.Format(layout))
				preSLA := spliteSLADataByEnv(resultData)
				calSLa(preSLA)

			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func spliteSLADataByEnv(resultData Result) ([]SLA) {
	var sla []SLA

	for _, v := range resultData.Data.Metrics {
		vStr := v.Value
		if vStr == "0" {
			log.Println("error,%s sla is 0", v.Name)
			continue
		}

		sla = append(sla, v)

	}
	return sla
}

func getSlaData(startDate string,endDate string) Result {
	var p GraphqlParam
	p.Query = "query queryData($condition: TopNCondition!, $duration: Duration!) {sortMetrics: sortMetrics(condition: $condition, duration: $duration) {name id value refId}}"
	p.Variables.Condition.Name = "service_sla"
	p.Variables.Condition.Normal = true
	p.Variables.Condition.Order = "ASC"
	p.Variables.Condition.Scope = "Service"
	p.Variables.Condition.TopN = "200"
	p.Variables.Condition.ParentService = ""
	p.Variables.Duration.Start = startDate
	p.Variables.Duration.End = endDate
	p.Variables.Duration.Step = "MINUTE"

	pByte, err := json.Marshal(p)

	if err != nil {
		fmt.Print("error ", err)
	}

	pStr := string(pByte[:])

	log.Println(pStr)

	resp, err := http.Post("http://sw8collector.res.eqxiu.com/graphql", "application/json;charset=UTF-8", strings.NewReader(pStr))
	if err != nil {
		log.Println(resp)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	//fmt.Println(string(body))

	var resultData Result
	err = json.Unmarshal(body, &resultData)
	if err != nil {
		log.Println("some error", err)
	}
	return resultData
}

func calSLa(list []SLA) {
	var count = 0
	var totalSla = 0.00

	for _, v := range list {
		vStr := v.Value
		slaFloat := 100.00
		if vStr != "10000" {
			slaStr := vStr
			if len(vStr) >2 {
				slaStr = vStr[0:2] + "." + vStr[3:]
			}
			slaFloat, _ = strconv.ParseFloat(slaStr, 32)
		}

		count++
		totalSla += slaFloat
		//fmt.Printf("%d %s sla=%f \n", count, v.Name, slaFloat)

		dubbo_service_sla.WithLabelValues(v.Name).Set(slaFloat)

	}

	fmt.Printf("count = %d \n", count)
	fmt.Printf("totalSla = %f \n", totalSla/float64(count))
}




