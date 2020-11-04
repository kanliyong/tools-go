package main

import (
	"fmt"
	"os/exec"
)

func main(){
	data, _ := exec.Command("bash","-c",`echo "db.devex_request.find().limit(1).pretty()" | mongo mongodb://192.168.0.18:27010/data-mgr`).Output()
	fmt.Println("%s",string(data))
}
