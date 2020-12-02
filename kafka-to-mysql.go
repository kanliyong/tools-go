package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/go-sql-driver/mysql"
	//"github.com/pkg/profile"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)
/**
docker build -t ccr.ccs.tencentyun.com/eqxiu/kafka-to-mysql -f kafka-to-mysql.Dockerfile .
docker push ccr.ccs.tencentyun.com/eqxiu/kafka-to-mysql



./kafaka-to-mysql -servers="hadoop104.eqxiu.com:9092,hadoop105.eqxiu.com:9092,hadoop106.eqxiu.com:9092" \
-group_id=kafka-to-mysql \
-topic="gateway_original_show" \
-dsn="root:Eqxiu@2019@tcp(10.0.10.49)/nginx?charset=utf8mb4&collation=utf8mb4_general_ci" \
-table="waf_show_log"

*/
func parseTime(input string) time.Time {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", input)
	if err != nil {
		return time.Now()
	}
	return t
}

var db *sql.DB
var (
	servers string
	group_id string
	topic string
	DSN string
	table string
)

func init(){

	flag.StringVar(&servers,
		"servers",
		"hadoop104.eqxiu.com:9092,hadoop105.eqxiu.com:9092,hadoop106.eqxiu.com:9092",
		"kafka bootstrap servers")
	flag.StringVar(&group_id, "group_id","group_id_test", "kafka consumer group id")
	flag.StringVar(&topic, "topic","topic_name", "kafka consumer topic")
	flag.StringVar(&DSN, "dsn","root:Eqxiu@2019@tcp(10.0.10.49)/nginx?charset=utf8mb4&collation=utf8mb4_general_ci", "mysql DSN")
	flag.StringVar(&table, "table","waf_show_log","table name")
	flag.Parse()

	var err error
	db, err = sql.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	//defer db.Close()
	//See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(80)
	db.SetMaxIdleConns(10)
}
func insertBatch(array []NginxLog) {
	sqlStr := fmt.Sprintf(`insert into %s 
    (
    create_datetime,
    host,
    method,
    uri,
    status,
    request_time,
    response_body_size,
    agent,
    referer,
    upstream_addr,
    upstream_status,
    tracker_user_id,
    remote_addr
    )
    values `,table)

	vals := []interface{}{}

	for _, l := range array {
		sqlStr += "(?,?,?,?,?,?,?,?,?,?,?,?,?),"
		vals = append(vals, l.Time, l.Host, l.Method, l.Uri, l.Code, l.RequestTime,
			l.ResponseBodySize, l.Agent, l.Referer, l.UpstreamAddr, l.UpstreamStatus,
			l.Tracker, l.RemoteIp)
	}
	//trim the last ,
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()


	//prepare the statement
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	//format all vals at once
	_, err = stmt.Exec(vals...)
	if err != nil {
		//panic(err)
		log.Println(err)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	log.Printf("insert %d", len(array))
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
	var array []NginxLog
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
			array = append(array, data)
			if len(array) > 1000 {
				log.Println(">1000")
				insertBatch(array[:])
				array = nil
			}
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
	//defer profile.Start(profile.MemProfile).Stop()

	consumer()
	//go consumer()
	//go consumer()
	//var w = sync.WaitGroup{}
	//w.Add(1)
	//log.Println("start ...")
	//w.Wait()
}
