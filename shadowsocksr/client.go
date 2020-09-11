package shadowsocksr

import (
	"errors"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Qingluan/merkur/shadowsocksr/obfs"
	"github.com/Qingluan/merkur/shadowsocksr/protocol"
	"github.com/Qingluan/merkur/shadowsocksr/ssr"
)

var (
	DefaultTimeout = 12 * time.Second
)

func NewSSRClient(u *url.URL) (*SSTCPConn, error) {
	query := u.Query()
	encryptMethod := query.Get("encrypt-method")
	encryptKey := query.Get("encrypt-key")
	cipher, err := NewStreamCipher(encryptMethod, encryptKey)
	if err != nil {
		log.Println("err in NewStreamCipher:", err)
		return nil, err
	}

	dialer := net.Dialer{
		Timeout:   DefaultTimeout,
		DualStack: true,
	}
	conn, err := dialer.Dial("tcp", u.Host)
	if err != nil {

		log.Println("err in tcp connect :", u.Host, err)
		return nil, err
	}

	ssconn := NewSSTCPConn(conn, cipher)
	if ssconn.Conn == nil || ssconn.RemoteAddr() == nil {
		return nil, errors.New("nil connection")
	}

	// should initialize obfs/protocol now
	rs := strings.Split(ssconn.RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(rs[1])

	ssconn.IObfs = obfs.NewObfs(query.Get("obfs"))
	obfsServerInfo := &ssr.ServerInfoForObfs{
		Host:   rs[0],
		Port:   uint16(port),
		TcpMss: 1460,
		Param:  query.Get("obfs-param"),
	}
	ssconn.IObfs.SetServerInfo(obfsServerInfo)
	ssconn.IProtocol = protocol.NewProtocol(query.Get("protocol"))
	protocolServerInfo := &ssr.ServerInfoForObfs{
		Host:   rs[0],
		Port:   uint16(port),
		TcpMss: 1460,
		Param:  query.Get("protocol-param"),
	}
	ssconn.IProtocol.SetServerInfo(protocolServerInfo)

	return ssconn, nil
}
