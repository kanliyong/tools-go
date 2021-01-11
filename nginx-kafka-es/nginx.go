package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

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
	TraceId          string
}

func parseTime(input string) time.Time {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", input)
	if err != nil {
		return time.Now()
	}
	return t
}

func maxLength(content string, maxLen int) string {
	asRunes := []rune(content)
	if len(asRunes) > maxLen {
		return string(asRunes[:maxLen])
	} else {
		return content
	}
}

func onMessage(e *kafka.Message) ([]byte, error) {

	s := string(e.Value)
	rs := strings.Split(s, "#|#")

	code, _ := strconv.Atoi(rs[8])
	requestTime, _ := strconv.ParseFloat(rs[11], 8)
	ResponseBodySize, _ := strconv.ParseFloat(rs[15], 8)

	TraceId := ""
	if len(rs) > 19 {
		TraceId = rs[19]
	}
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
		TraceId:          TraceId,
	}
	return json.Marshal(data)
}
