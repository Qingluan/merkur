package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Qingluan/merkur"
	"github.com/cheggaaa/pb"
)

func main() {
	var testurlorder string
	var url string
	flag.StringVar(&testurlorder, "o", "", "order url or ssr/ss uri")
	flag.StringVar(&url, "u", "", "target test")
	flag.Parse()

	if url == "" {
		url = "https://www.google.com"
	}
	if testurlorder != "" && strings.HasPrefix(testurlorder, "http") {

		proxyPool := merkur.NewProxyPool(testurlorder)
		bar := pb.New(proxyPool.Count())

		for k, v := range proxyPool.LoopOneTurn(func(proxyDialer merkur.Dialer) interface{} {
			client := proxyDialer.ToHttpClient()
			if res, err := client.Get(url); err != nil {
				return err
			} else {
				return res.StatusCode
			}
		}, bar) {
			fmt.Println(k[:10], v)
		}

	} else {
		if client := merkur.NewProxyHttpClient(testurlorder); client != nil {
			st := time.Now()
			if res, err := client.Get(url); err != nil {
				conf, ierr := merkur.ParseUri(testurlorder)
				fmt.Println(conf, ierr)
				log.Println("used:", time.Now().Sub(st), "err:", err)
			} else {
				c, e := merkur.ParseUri(testurlorder)
				if e != nil {
					log.Println(e)
				}
				log.Println("used:", time.Now().Sub(st), "code:", res.StatusCode, "proxy:", c.Server)

			}
		}
	}
}
