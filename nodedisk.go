package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type host struct {
	Name   string `json:"name"`
	Ip     string `json:"ip"`
	User   string `json:"user"`
	Credit string `json:"credit"`
}

func printDisk(host host) {
	args := []string{
		"-i",
		host.Credit,
		fmt.Sprintf("%s@%s", host.User, host.Ip),
		"df -h",
	}
	data, err := exec.Command("ssh", args...).Output()
	if err != nil {
		log.Panicf("err=%v,host=%v", err, host)
	}
	fmt.Println(host.Ip)
	scanner := bufio.NewReader(bytes.NewReader(data))
	for {
		line, _, err := scanner.ReadLine()
		if err != nil && err == io.EOF {
			break
		}

		re := regexp.MustCompile(`/dev/vd`)
		if re.Match(line) {
			fmt.Println("\t", string(line))
		}
	}

}
func main() {
	for _, arg := range os.Args[1:] {
		hosts, _ := openFile(arg)
		for _, v := range hosts {
			printDisk(v)
		}
	}
}

func openFile(arg string) ([]host, error) {
	file, err := os.Open(arg)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var hosts []host
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &hosts)

	return hosts, err
}
