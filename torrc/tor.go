package torrc

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
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
	LocalService = TestIfServiceStartLocal()
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

func Socks5Dialer(addr string) (proxy.Dialer, error) {
	if strings.HasPrefix(addr, "socks5") {
		addr = strings.SplitN(addr, "://", 2)[1]
	}
	dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
	}
	return dialer, err
}

func NewTorDialer() (dialer proxy.Dialer, err error) {
	if LocalService {
		return Socks5Dialer("localhost:9050")
	}

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

func TestIfServiceStartLocal() bool {
	if _, err := os.Stat("/etc/tor/torrc"); err == nil {
		c, err := net.Dial("tcp", "localhost:9050")
		if err != nil {
			return false
		} else {
			defer c.Close()
			return true
		}
	}
	return false
}
