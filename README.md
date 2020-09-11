

## Usage

```golang
// 
//  support vmess / ss / ssr / socks 
//

if dialer := NewProxyDialer("vmess://..."); dialer != nil{
        dialer.Dial("tcp",target)
}


// http:

if client := merkur.NewProxyHttpClient("ssr://...."); client != nil {
        client.Get("https://www.google.com")
}


```

## Example

```golang


package main

import (
        "flag"
        "fmt"
        "log"
        "strings"
        "sync"
        "time"

        "github.com/Qingluan/merkur"
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
        var wait sync.WaitGroup
        if testurlorder != "" && strings.HasPrefix(testurlorder, "http") {
                for _, ssurl := range merkur.ParserOrder(testurlorder) {
                        fmt.Println(ssurl)
                        wait.Add(1)
                        go func(url string, ssurl string) {
                                defer wait.Done()
                                if client := merkur.NewProxyHttpClient(ssurl); client != nil {
                                        st := time.Now()
                                        if res, err := client.Get(url); err != nil {
                                                log.Println("used:", time.Now().Sub(st), "err:", err)
                                        } else {
                                                log.Println("used:", time.Now().Sub(st), "code:", res.StatusCode, "proxy:", merkur.MustParseUri(ssurl))
                                        }
                                }

                        }(url, ssurl)
                }
                wait.Wait()
        } else {
                if client := merkur.NewProxyHttpClient(testurlorder); client != nil {
                        st := time.Now()
                        if res, err := client.Get(url); err != nil {
                                log.Println("used:", time.Now().Sub(st), "err:", err)
                        } else {
                                log.Println("used:", time.Now().Sub(st), "code:", res.StatusCode, "proxy:", merkur.MustParseUri(testurlorder))
                        }
                }
        }
}


```
