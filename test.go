package main

import "regexp"

func main(){
	re := regexp.MustCompile("10.0.6.16:9104")
	println(re.MatchString("10.0.6.16:9104"))
}
