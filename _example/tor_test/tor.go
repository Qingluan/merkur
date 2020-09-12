package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Qingluan/merkur"
)

func main() {
	url := "http://ifconfig.co/json"

	client := merkur.NewProxyHttpClient("tor://")
	client.Get(url)
	res, err := client.Get(url)
	if err != nil {
		// panic(err)
		log.Println("Get err:", err)
		return
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf))
}
