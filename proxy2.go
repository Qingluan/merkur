package merkur

import (
	"errors"
	"log"
	"strings"

	"github.com/Qingluan/merkur/config"
	// logger "github.com/isayme/go-logger"
	// ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

var (
	ParseOrder = config.ParseOrding
	ParseUri   = func(uri string) (cfg config.Config, err error) {
		if strings.HasPrefix(uri, "vmess://") {
			return config.ParseVmessUri(uri)
		} else if strings.HasPrefix(uri, "ss") {
			return config.ParseSSUri(uri)
		} else {
			return cfg, errors.New("Invalid uri:" + uri)
		}
	}
	MustParseUri = func(uri string) (cfg config.Config) {
		var e error
		if strings.HasPrefix(uri, "vmess://") {
			cfg, e = config.ParseVmessUri(uri)
		} else if strings.HasPrefix(uri, "ss") {
			cfg, e = config.ParseSSUri(uri)
		} else {
			return cfg
		}
		if e != nil {
			log.Fatal("parse uri:", e, uri)
		}
		return cfg
	}
)
