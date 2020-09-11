package merkur

import (
	"crypto/tls"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	DefaultTimeout = 12
)

func ChooiceOne(array []interface{}) interface{} {
	all := len(array)
	i := rand.Int() % all
	return array[i]
}

func (proxy *ProxyDialer) ToHttpClient(timeout ...int) (client *http.Client) {
	tout := time.Second * time.Duration(DefaultTimeout)
	if timeout != nil {
		tout = time.Second * time.Duration(timeout[0])
	}
	transport := http.Transport{
		ResponseHeaderTimeout: tout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	transport.Dial = proxy.Dial

	client = &http.Client{
		Transport: &transport,
		Timeout:   tout,
	}
	return
}

func NewProxyHttpClient(proxy interface{}, timeout ...int) (client *http.Client) {
	tout := time.Second * time.Duration(DefaultTimeout)
	if timeout != nil {
		tout = time.Second * time.Duration(timeout[0])
	}
	transport := http.Transport{
		ResponseHeaderTimeout: tout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	if dialer := NewProxyDialer(proxy); dialer != nil {
		transport.Dial = dialer.Dial
	} else {
		log.Println("parse proxy err:", proxy)
		return
	}

	client = &http.Client{
		Transport: &transport,
		Timeout:   tout,
	}
	return
}
