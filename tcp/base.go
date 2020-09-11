package tcp

import (
	"net"

	"github.com/Qingluan/merkur/config"
)

type DefaultTcpDialer struct {
}

func (tcpd DefaultTcpDialer) Dial(nework, addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, config.Timeout)
}
