package main

import (
	"net/url"
	"testing"
)

func TestUrl(t *testing.T) {
	s := url.QueryEscape("<waf_nginx_log_{now/d}>")
	if s != "%3Cwaf_nginx_log_%7Bnow%2Fd%7D%3E" {
		t.Error(s)
	}
}

func TestSilce(t *testing.T) {
	s := []int{1, 2, 4}
	println(s[20])
}
