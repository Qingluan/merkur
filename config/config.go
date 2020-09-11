package config

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	IPMATCH      = regexp.MustCompile(`[\d\.]+`)
	InvalidSSURI = errors.New("Not Validat SSURI ")
	Timeout      = time.Second * 10
)

type Config struct {
	Server        string `json:"server"`
	Password      string `json:"password"`
	Method        string `json:"method"`
	ServerPort    int    `json:"server_port"`
	Timeout       int    `json:"timeout"`
	LocalPort     int    `json:"local_port"`
	LocalAddress  string `json:"local_address"`
	Protocol      string `json:"protocol"`
	Obfs          string `json:"obfs"`
	ConfigType    string `json:"conf_type"`
	ObfsParam     string `json:"obfs-param"`
	ProtocolParam string `json:"protocol-param"`
	OptionID      int    `json:"aid"`
	OptUUID       string `json:"uuid"`
}

func ToConfig(uri string) (cfg Config) {
	cfg, _ = ParseSSUri(uri)
	return
}

func (config Config) String() string {
	f, _ := json.MarshalIndent(&config, "", "\t")
	return string(f)
}

// PackAddr : pack to    || 1 (ipv4) | ip(4) | port(2) ||  or || 1 (domain) | 1 (len) | n (host str bytes) | 2 (port) ||
func PackAddr(addr string) []byte {
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}
	fs := strings.SplitN(addr, ":", 2)
	port, _ := strconv.Atoi(fs[1])
	buffer := bytes.NewBuffer([]byte{})
	// net.ResolveIPAddr()
	if IPMATCH.Match([]byte(fs[0])) && strings.Count(fs[0], ".") == 3 {
		buffer.Write([]byte{0x1})
		// fss := strings.SplitN(fs[0], ".", 4)
		ip := net.ParseIP(fs[0])
		buffer.Write([]byte(ip))
	} else {
		domainLen := len(fs[0])
		buffer.Write([]byte{0x3, byte(domainLen)})
		buffer.Write([]byte(fs[0]))
	}
	portb := make([]byte, 2)
	binary.BigEndian.PutUint16(portb, uint16(port))
	buffer.Write(portb)
	return buffer.Bytes()
}

func b64decode(a string) (o string, err error) {
	var ierr error
	if strings.Contains(a, "_") {
		a = strings.ReplaceAll(a, "_", "/")
	}

	if strings.Contains(a, "-") {
		a = strings.ReplaceAll(a, "-", "+")
	}

	for i := 0; i < 4; i++ {
		dat, err := base64.StdEncoding.DecodeString(strings.TrimSpace(a))
		if err != nil {
			a += "="
			ierr = err
			continue
		}
		return string(dat), nil
	}
	return "", ierr
}

func ParseVmessUri(u string) (cfg Config, err error) {
	var dats string
	if strings.HasPrefix(u, "vmess://") {
		u = u[8:]
	}
	if dats, err = b64decode(u); err != nil {
		return
	}
	s := make(map[string]interface{})
	err = json.Unmarshal([]byte(dats), &s)

	if err != nil {
		log.Println("json parse err:", err, u)

		return
	}
	// fmt.Println(s)
	if host, ok := s["host"]; ok {
		cfg.Server = host.(string)
	}

	if netAddr, ok := s["add"]; ok {
		cfg.ObfsParam = netAddr.(string)
	}

	if ports, ok := s["port"]; ok {
		cfg.ServerPort = int(ports.(float64))

	}
	if aids, ok := s["aid"]; ok {
		cfg.OptionID = int(aids.(float64))

	}
	if proto, ok := s["net"]; ok {
		cfg.Protocol = proto.(string)
		if cfg.Protocol == "ws" {
			cfg.ProtocolParam = s["path"].(string)
		}
	}

	if uid, ok := s["id"]; ok {
		cfg.OptUUID = uid.(string)
	}

	if sectype, ok := s["type"]; ok {
		cfg.Obfs = sectype.(string)
	}
	cfg.ConfigType = "vmess"
	return
}

func parseSSR(u string) (cfg Config, err error) {
	if strings.HasPrefix(u, "ssr://") {
		u = u[6:]
	}
	var dats string
	if dats, err = b64decode(u); err != nil {
		log.Println("err base64: >>", u, "<<")
		return
	}
	parts := strings.Split(dats, ":")
	cfg.Server = parts[0]
	cfg.ServerPort, _ = strconv.Atoi(parts[1])
	cfg.Protocol = parts[2]
	cfg.Method = parts[3]
	cfg.Obfs = parts[4]
	cfg.ConfigType = "ssr"
	parts_pwd := strings.SplitN(parts[5], "/", 2)

	if cfg.Password, err = b64decode(parts_pwd[0]); err != nil {
		log.Println("err base64: >>", u, "<<", " >>>", parts_pwd[0], "<<<")
		return
	}
	ur, _ := url.Parse("ssr://abc/" + parts_pwd[1])
	values := ur.Query()
	if cfg.ObfsParam, err = b64decode(values.Get("obfsparam")); err != nil {
		return
	}
	if cfg.ProtocolParam, err = b64decode(values.Get("protoparam")); err != nil {
		return
	}
	return
}

func ParseSSUri(u string) (cfg Config, err error) {
	cfg = Config{}
	if u == "" {
		return
	}
	// invalidURI := errors.New("invalid URI")
	// ss://base64(method:password)@host:port
	// ss://base64(method:password@host:port)
	if strings.HasPrefix(u, "ssr://") {
		u = strings.TrimLeft(u, "ssr://")
		return parseSSR(u)
	} else if strings.HasPrefix(u, "ss://") {
		u = strings.TrimLeft(u, "ss://")
	} else if strings.HasPrefix(u, "vmess://") {
		u = u[8:]
		// fmt.Println("\n||", u)
		return ParseVmessUri(u)
	}
	i := strings.IndexRune(u, '@')
	var headParts, tailParts [][]byte
	// var err error
	if i == -1 {
		var dats string
		if dats, err = b64decode(u); err != nil {
			return
		}
		dat := []byte(dats)
		parts := bytes.Split(dat, []byte("@"))
		if len(parts) != 2 {
			log.Println("lack @ e")
			return
		}
		headParts = bytes.SplitN(parts[0], []byte(":"), 2)
		tailParts = bytes.SplitN(parts[1], []byte(":"), 2)

	} else {
		if i+1 >= len(u) {

			return
		}
		tailParts = bytes.SplitN([]byte(u[i+1:]), []byte(":"), 2)
		dat, ierr := base64.StdEncoding.DecodeString(u[:i])
		if ierr != nil {
			return cfg, ierr
		}
		headParts = bytes.SplitN(dat, []byte(":"), 2)
	}
	if len(headParts) != 2 {
		return
	}

	if len(tailParts) != 2 {
		return
	}
	cfg.Method = string(headParts[0])

	cfg.Password = string(headParts[1])
	p, e := strconv.Atoi(string(tailParts[1]))
	if e != nil {
		return
	}
	cfg.Server = string(tailParts[0])
	cfg.ServerPort = p
	return
}

func ParseOrding(urlOrbuf string) (ssuri []string) {
	raw := []byte{}
	if strings.HasPrefix(urlOrbuf, "http") {
		log.Println("download all proxy configs : ", urlOrbuf)
		if res, err := http.Get(urlOrbuf); err == nil {
			// var raw []byte
			switch res.Header.Get("Content-Encoding") {
			case "gzip":
				reader, err := gzip.NewReader(res.Body)
				if err != nil {
					log.Println("parse body gzip data error:", err)
					return nil
				}
				defer reader.Close()
				raw, _ = ioutil.ReadAll(reader)
			default:
				raw, _ = ioutil.ReadAll(res.Body)
			}
		} else {
			log.Println("no proxy config in url :", urlOrbuf)
		}
	} else {
		raw = []byte(urlOrbuf)
	}
	de, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		log.Println("invalid ording data!")
		return
	}
	for _, uri := range bytes.Split(de, []byte("\n")) {
		one := strings.TrimSpace(string(uri))
		if strings.Contains(one, "://") {
			ssuri = append(ssuri, one)
		}
	}
	return
}
