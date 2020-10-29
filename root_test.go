package main

import (
	"testing"
	"tools/cmd"
)

// func TestScaleZero(t *testing.T) {
// 	claster := cmd.ClusterArray[0]
// 	cmd.ScaleZero(claster, "iom-max-vue-testenv-1603353743")
// }

func TestMongo(t *testing.T) {
	cmd.RmMongo("iom-max-vue")
}
