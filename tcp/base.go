package tcp

import "net"

type DefaultTcpDialer struct {
}

func (tcpd DefaultTcpDialer) Dial(nework, addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}
