package torrc

import (
	"context"
	"fmt"
	"time"

	"github.com/cretz/bine/tor"
	"golang.org/x/net/proxy"
)

var (
	TorService = struct {
		State   int
		Service *tor.Tor
	}{
		0,
		nil,
	}
)

func IsStart() bool {
	if TorService.State == 0 || TorService.Service == nil {
		return false
	}
	return true
}

func StartTor() (err error) {
	defer func() {
		TorService.State = 1
	}()
	// if path, err := exec.LookPath("tor");err == nil{
	// 	if _, err := os.Stat("/etc/torrc")
	// }
	fmt.Println("Starting tor and fetching title of https://check.torproject.org, please wait a few seconds...")
	// conf := &tor.StartConf{}
	TorService.Service, err = tor.Start(nil, nil)
	return
}

func NewTorDialer() (dialer proxy.Dialer, err error) {
	if !IsStart() {
		if err = StartTor(); err != nil {
			return
		}
	}
	// defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, _ := context.WithTimeout(context.Background(), time.Minute)
	dialer, err = TorService.Service.Dialer(dialCtx, nil)
	return
}
