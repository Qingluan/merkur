package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/merkur"
	"github.com/cheggaaa/pb"
)

func main() {
	var testurlorder string
	var url string
	var testflow bool
	flag.StringVar(&testurlorder, "o", "", "order url or ssr/ss uri")
	flag.StringVar(&url, "u", "", "target test")
	flag.BoolVar(&testflow, "f", false, "test flow order")

	flag.Parse()

	if url == "" {
		url = "http://ifconfig.co/json"
	}
	if testurlorder != "" && strings.HasPrefix(testurlorder, "http") {
		if testflow {
			var wait sync.WaitGroup
			for i := 0; i < 10; i++ {
				wait.Add(1)

				// go func(client *http.Client) {

				go func() {
					client := merkur.NewProxyHttpClient(testurlorder)

					defer wait.Done()
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
				}()
				// }(client)

			}
			wait.Wait()
			os.Exit(0)
		}

		proxyPool := merkur.NewProxyPool(testurlorder)
		bar := pb.New(proxyPool.Count())

		for k, v := range proxyPool.LoopOneTurn(func(proxyDialer merkur.Dialer) interface{} {
			client := proxyDialer.ToHttpClient()
			if res, err := client.Get(url); err != nil {
				return err
			} else {
				return res
			}
		}, bar) {
			switch v.(type) {
			case error:
				log.Println(v.(error))
				log.Println("conf : ", k)
				c, _ := merkur.ParseUri(k)
				log.Println(c)
			case *http.Response:
				res := v.(*http.Response)
				r, _ := json.MarshalIndent(res.Header, "", "\t")
				// c, _ := merkur.ParseUri(k)
				fmt.Println(res.StatusCode, string(r), k)

			}

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
