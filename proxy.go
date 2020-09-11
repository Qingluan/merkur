package merkur

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Qingluan/merkur/config"
	"github.com/Qingluan/merkur/shadowsocks/cipher"
	"github.com/Qingluan/merkur/shadowsocksr"
	"github.com/Qingluan/merkur/tcp"
	"github.com/Qingluan/merkur/vmess"
	"github.com/Qingluan/merkur/ws"
	"golang.org/x/net/proxy"
)

var (
	DefaultProxyPool = NewProxyPool()
)

type Dialer interface {
	Dial(network, addr string) (net.Conn, error)
	ToHttpClient(timeout ...int) (client *http.Client)
	ProxyURI() string
}

type ProxyDialer struct {
	url  string
	conf config.Config
}

func (dialer *ProxyDialer) ProxyURI() string {
	return dialer.url
}

func NewDialer(ssconf config.Config) (dialer *ProxyDialer, err error) {
	dialer = new(ProxyDialer)
	dialer.conf = ssconf
	dialer.url = fmt.Sprintf("%s:%d", ssconf.Server, ssconf.ServerPort)
	if dialer.conf.Server == "" {
		err = config.InvalidSSURI
	}
	return
}

func NewDialerByURI(ssuri string) (dialer *ProxyDialer, err error) {
	dialer = new(ProxyDialer)
	dialer.conf, err = config.ParseSSUri(ssuri)
	dialer.url = ssuri
	if err != nil {
		return
	}
	// fmt.Println(dialer.conf)
	if dialer.conf.Server == "" {
		err = config.InvalidSSURI
	}
	return
}

func NewProxyDialer(proxyObj interface{}) (dialer proxy.Dialer) {
	switch proxyObj.(type) {
	case proxy.Dialer:
		return proxyObj.(proxy.Dialer)
	case *ProxyPool:
		DefaultProxyPool.Merge(*proxyObj.(*ProxyPool))
		uri := DefaultProxyPool.Get()
		dial, err := NewDialerByURI(uri)
		if err != nil {
			log.Println("failed use ss proxy dialer:", err)
			return
		}
		return dial
	case ProxyPool:
		DefaultProxyPool.Merge(proxyObj.(ProxyPool))
		uri := DefaultProxyPool.Get()
		dial, err := NewDialerByURI(uri)
		if err != nil {
			log.Println("failed use ss proxy dialer:", err)
			return
		}
		return dial
	case *config.Config:
		dial, err := NewDialer(*proxyObj.(*config.Config))
		if err != nil {
			log.Println("failed use conf proxy dialer:", err)
			return
		}
		return dial
	case config.Config:
		dial, err := NewDialer(proxyObj.(config.Config))
		if err != nil {
			log.Println("failed use conf proxy dialer:", err)
			return
		}
		return dial
	case string:
		if strings.HasPrefix(proxyObj.(string), "ss") {

			dial, err := NewDialerByURI(proxyObj.(string))
			if err != nil {
				log.Println("failed use ss proxy dialer:", err)
				return
			}
			return dial
		} else if strings.HasPrefix(proxyObj.(string), "vmess://") {
			dial, err := NewDialerByURI(proxyObj.(string))
			if err != nil {
				log.Println("failed use vmess proxy dialer:", err)
				return
			}
			return dial
		} else if strings.HasPrefix(proxyObj.(string), "socks5://") {
			dialer := Socks5Dialer(proxyObj.(string))
			return dialer
		} else if strings.HasPrefix(proxyObj.(string), "http") {
			DefaultProxyPool.Add(proxyObj.(string))
			return DefaultProxyPool.GetDialer()
		}
	}
	return

}

func (ss *ProxyDialer) Dial(network string, addr string) (con net.Conn, err error) {

	// CipherWith(ss.conf.Method, ss.conf.Password, func(sscipher *cipher.Cipher){
	switch ss.conf.ConfigType {
	case "vmess":
		var predial proxy.Dialer
		if ss.conf.Protocol == "ws" {
			wsAddr := fmt.Sprintf("ws://%s:%d%s", ss.conf.ObfsParam, ss.conf.ServerPort, ss.conf.ProtocolParam)
			predial, err = ws.NewWSDialer(wsAddr, nil)
			if err != nil {
				return
			}
		} else {
			predial = tcp.DefaultTcpDialer{}
		}
		vmessDialer, err := vmess.NewVMessDialer(fmt.Sprintf("%s:%d", ss.conf.Server, ss.conf.ServerPort), ss.conf.OptUUID, ss.conf.Obfs, ss.conf.OptionID, predial)
		if err != nil {
			return nil, err
		}
		con, err = vmessDialer.Dial(network, addr)
		if err != nil && ss.conf.ObfsParam != ss.conf.Server {
			// log.Println("dial err:", err, "try again")
			if ss.conf.Protocol == "ws" {
				wsAddr := fmt.Sprintf("ws://%s:%d%s", ss.conf.Server, ss.conf.ServerPort, ss.conf.ProtocolParam)
				predial, err = ws.NewWSDialer(wsAddr, nil)
				if err != nil {
					return nil, err
				}
			} else {
				predial = tcp.DefaultTcpDialer{}
			}
			vmessDialer, err := vmess.NewVMessDialer(fmt.Sprintf("%s:%d", ss.conf.ObfsParam, ss.conf.ServerPort), ss.conf.OptUUID, ss.conf.Obfs, ss.conf.OptionID, predial)
			if err != nil {
				return nil, err
			}
			con, err = vmessDialer.Dial(network, addr)
		}
	case "ssr":
		u := &url.URL{
			Scheme: "ssr",
			Host:   fmt.Sprintf("%s:%d", ss.conf.Server, ss.conf.ServerPort),
		}
		v := u.Query()
		v.Set("encrypt-method", ss.conf.Method)
		v.Set("encrypt-key", ss.conf.Password)
		v.Set("obfs", ss.conf.Obfs)
		v.Set("obfs-param", ss.conf.ObfsParam)
		v.Set("protocol", ss.conf.Protocol)
		v.Set("protocol-param", ss.conf.ProtocolParam)

		u.RawQuery = v.Encode()

		ssrconn, err := shadowsocksr.NewSSRClient(u)
		if err != nil {
			// Info(ss.conf)
			return nil, fmt.Errorf("connecting to SSR server failed :%v", err)
		}

		// if bi.ObfsData == nil {
		// 	bi.ObfsData =
		// }
		ssrconn.IObfs.SetData(ssrconn.IObfs.GetData())
		ssrconn.IProtocol.SetData(ssrconn.IProtocol.GetData())

		if _, err := ssrconn.Write(config.PackAddr(addr)); err != nil {
			ssrconn.Close()
			return nil, err
		}
		return ssrconn, nil
	default:
		ssconn, ierr := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ss.conf.Server, ss.conf.ServerPort), 5*time.Second)
		if ierr != nil {
			return nil, ierr
		}
		ssconfig := ss.conf
		sscipher := cipher.NewCipher(ssconfig.Method)
		key := cipher.NewKey(ssconfig.Method, ssconfig.Password)
		sscipher.Init(key, ssconn)
		con = sscipher
		if _, err = con.Write(config.PackAddr(addr)); err != nil {
			con.Close()
			return nil, err
		}
		return

	}
	return
}

func Socks5Dialer(addr string) proxy.Dialer {
	if strings.HasPrefix(addr, "socks5") {
		addr = strings.SplitN(addr, "://", 2)[1]
	}
	dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
	}
	return dialer
}
