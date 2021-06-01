package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Qingluan/merkur"
	"github.com/Qingluan/merkur/socks5"

	// "github.com/martinlindhe/notify"
	"golang.org/x/net/proxy"
)

var (
	TO_STOP               = false
	RE_START              = 0
	NowConfig             = ""
	Socks5ConnectedRemote = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43}
	Configs               = Init()
	DefaultPool           = merkur.DefaultProxyPool
)

type ConfigK struct {
	Routes     map[string]string `json:"routes"`
	ListenHost string            `json:"listen"`
}

func Init() ConfigK {
	return ConfigK{
		Routes:     make(map[string]string),
		ListenHost: "0.0.0.0:1080",
	}
}

func Listen(testUrl string) (err error) {
	ln, err := net.Listen("tcp", Configs.ListenHost)
	ln2, err2 := net.Listen("tcp", Configs.ListenHost+"0")
	if err2 != nil {
		log.Fatal(err2)
	}
	go func() {

		for {
			p, err := ln2.Accept()
			log.Println("try add new proxy.....")
			if err != nil {
				log.Fatal(err)
				break
			}
			buf := make([]byte, 1024)
			p.SetDeadline(time.Now().Add(10 * time.Second))
			if n, err := p.Read(buf); err == nil {
				log.Println("Add :", string(buf))
				DefaultPool.Add(strings.TrimSpace(string(buf[:n])))
				if strings.HasPrefix(string(buf), "http") {
					DefaultPool.TestAll(testUrl)
				}
			}
		}
	}()

	// if conn.ShowLog < 2 {
	// 	// utils.ColorL("Local Listen:", listenAddr)

	// }
	// dialer := merkur.NewProxyDialer(NowConfig)
	log.Println("Set proxy:", NowConfig, " Listen:", Configs.ListenHost)

	// log.Println("FrameV2", "Start", Configs.ListenHost, "")

	for {
		if TO_STOP {
			break
		}
		// if conn.Role == "tester" && conn.GetAliveNum() > conn.Numconn {
		// 	time.Sleep(10 * time.Millisecond)
		// 	continue
		// }
		p1, err := ln.Accept()

		if err != nil {
			if !strings.Contains(err.Error(), "too many open files") {
				LogErr(err)
			}

			continue
		}
		go handleSocks5TcpAndUDP(p1, nil)

	}
	// if RE_START > 0 {
	// 	Listen()
	// }
	return
}

func handleSocks5TcpAndUDP(p1 net.Conn, dialer proxy.Dialer) {
	defer p1.Close()
	if err := socks5.Socks5HandShake(&p1); err != nil {
		// utils.ColorL("socks handshake:", err)
		return
	}

	_, host, _, err := socks5.GetLocalRequest(&p1)
	if err != nil {
		LogErr(err)
		return
	}
	// fmt.Println(string(raw))
	// if isUdp {

	// utils.ColorL("socks5 UDP-->", host)
	// } else {

	// log.Println("socks5 -->", host)
	// }
	if err != nil {
		LogErr(err)
		return
	}
	handleBody(p1, dialer, host)
}

func LogErr(err error) bool {
	if err != nil {
		// notify.Alert("FrameV2", "Error info:", err.Error(), "")
		return true
	}
	return false
}

func handleBody(p1 net.Conn, dialer proxy.Dialer, host string) {
	if NowConfig != "" {
		if dialer == nil {
			dialer = DefaultPool.GetDialer()
		}
		p2, err := dialer.Dial("tcp", host)
		// log.Println("connecting ->", host)
		if err != nil {
			log.Println("Err", "error", err.Error(), "")
			return
		}
		if p2 == nil {
			// log.Println("Frame2", "error", "p2 is not connected", "")
			return
		}
		if p1 != nil && p2 != nil {
			// log.Println("connect ->", host)
			_, err = p1.Write(Socks5ConnectedRemote)
			if err != nil {
				LogErr(err)
				return
			}
			// log.Println("connected ->", host)
			// PipeOne(p2, p1)
			socks5.Pipe(p1, p2)
		} else {
			log.Println("Frame2", "err", "connection is not connected!", "")
		}
	} else {
		// log.Println("Frame2", "error", "No set Proxy item", "")
		return
	}
}

func main() {
	server := ""
	proxyURI := ""
	addUrl := ""
	testURL := ""
	flag.StringVar(&server, "l", "0.0.0.0:1080", "set liseten address.")
	flag.StringVar(&proxyURI, "p", "", "can vmess:// | ss:// | ssr:// | order address (ssr/v2ray/ss) http https://somevps.com/order ")
	flag.StringVar(&addUrl, "a", "", "add order /ss/ vmess / ssr ")
	flag.StringVar(&testURL, "t", "https://www.google.com", "set a url to test ")
	flag.Parse()

	if addUrl != "" {
		t, e := net.Dial("tcp", "localhost:10800")
		if e != nil {
			log.Println(e)
			return
		}
		t.Write([]byte(addUrl))
		os.Exit(0)
	}

	NowConfig = proxyURI
	Configs.ListenHost = server
	DefaultPool.Add(proxyURI)
	DefaultPool.TestAll(testURL)
	if err := Listen(testURL); err != nil {
		log.Fatal(err)
	}
}
