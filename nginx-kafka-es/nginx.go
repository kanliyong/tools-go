package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

/**


{
	remote_addr,
	time_local ,
	jsessionid,
	host,
	request_method,
	uri,
	args,
	request_body,
	status,
	http_user_agent,
	http_referer,
	request_time,
	upstream_addr,
	upstream_status,
	upstream_response_time,
	body_bytes_sent,
	http_x_forwarded_for,
	upstream_http_set_cookie,
	tracker_user_id,
	request_id
}

180.139.200.118#|#
29/Jan/2021:14:30:15 +0800#|#
58f6d52c1e734626ad0edc2e6a572e59#|#
work-api.eqxiu.com#|#
GET#|#
/m/market/content/myScene#|#
groupId=&pageNo=1&pageSize=23&sceneName=#|#
-#|#
200#|#
Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36#|#
http://www.eqxiu.com/eip/scene?type=H5&show_tk_id=show-pro-711-711-0-0-0#|#
0.178#|#
10.0.6.144:80#|#
200#|#
0.171#|#
3114#|#
180.139.200.118#|#
-#|#
235e4fef2ab64e288b36d1b5ea350fc8
*/
type NginxLog struct {
	Remote_addr              string
	Time_local               time.Time
	Jsessionid               string
	Host                     string
	Request_method           string
	Uri                      string
	Args                     string
	Request_body             string
	Status                   int
	Http_user_agent          string
	Http_referer             string
	Request_time             float64
	Upstream_addr            string
	Upstream_status          int
	Upstream_response_time   float64
	Body_bytes_sent          float64
	Http_x_forwarded_for     string
	Upstream_http_set_cookie string
	Tracker_user_id          string
	Request_id               string
	SWURI                    string `json:"sw_url"`
}

func parseTime(input string) time.Time {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", input)
	if err != nil {
		return time.Now()
	}
	return t
}

func parseInt(input string) int {
	code, err := strconv.Atoi(input)
	if err != nil {
		return 0
	} else {
		return code
	}
}

func parseFloat(input string) float64 {
	f, err := strconv.ParseFloat(input, 8)
	if err != nil {
		return 0
	} else {
		return f
	}
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

	for {
		if len(rs) < 20 {
			rs = append(rs, "")
		} else {
			break
		}
	}

	data := NginxLog{
		Remote_addr:              rs[0],
		Time_local:               parseTime(rs[1]),
		Jsessionid:               rs[2],
		Host:                     rs[3],
		Request_method:           rs[4],
		Uri:                      rs[5],
		Args:                     rs[6],
		Request_body:             rs[7],
		Status:                   parseInt(rs[8]),
		Http_user_agent:          rs[9],
		Http_referer:             rs[10],
		Request_time:             parseFloat(rs[11]),
		Upstream_addr:            rs[12],
		Upstream_status:          parseInt(rs[13]),
		Upstream_response_time:   parseFloat(rs[14]),
		Body_bytes_sent:          parseFloat(rs[15]),
		Http_x_forwarded_for:     rs[16],
		Upstream_http_set_cookie: rs[17],
		Tracker_user_id:          rs[18],
		Request_id:               rs[19],
		SWURI:                    "https://sw8.eff.eqxiu.cc/trace?traceId=" + rs[19],
	}
	return json.Marshal(data)
}
