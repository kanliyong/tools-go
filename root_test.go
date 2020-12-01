package main

import (
	"testing"
)

// func TestScaleZero(t *testing.T) {
// 	claster := cmd.ClusterArray[0]
// 	cmd.ScaleZero(claster, "iom-max-vue-testenv-1603353743")
// }

func TestMongo(t *testing.T) {
	tt := parseTime("10/Jun/2020:23:55:50 +0800")
	println(tt)
}
