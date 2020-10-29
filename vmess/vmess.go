package vmess

import (
	"errors"
	"log"
	"net"

	"github.com/Qingluan/merkur/tcp"
	"golang.org/x/net/proxy"
	// "golang.org/x/net/proxy"
	// "github.com/nadoo/glider/proxy"
)

// VMess struct.
type VMess struct {
	dialer proxy.Dialer
	addr   string

	uuid     string
	alterID  int
	security string

	client *Client
}

// NewVMess returns a vmess proxy.
func NewVMess(remote, uuid, sectp string, aid int, d proxy.Dialer) (*VMess, error) {

	addr := remote
	security := sectp

	client, err := NewClient(uuid, security, aid)
	if err != nil {
		log.Println("create vmess client err: %s", err)
		return nil, err
	}
	if d == nil {
		d = tcp.DefaultTcpDialer{}
	}
	p := &VMess{
		dialer:   d,
		addr:     addr,
		uuid:     uuid,
		alterID:  aid,
		security: security,
		client:   client,
	}

	return p, nil
}

// NewVMessDialer returns a vmess proxy dialer.
func NewVMessDialer(remote, uuid, sectp string, aid int, dialer proxy.Dialer) (proxy.Dialer, error) {
	return NewVMess(remote, uuid, sectp, aid, dialer)
}

// Addr returns forwarder's address.
func (s *VMess) Addr() string {
	if s.addr == "" {
		return ""
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the proxy.
func (s *VMess) Dial(network, addr string) (net.Conn, error) {
	// log.Println(s)
	rc, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		return nil, err
	}

	return s.client.NewConn(rc, addr)
}

// DialUDP connects to the given address via the proxy.
func (s *VMess) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	return nil, nil, errors.New("vmess client does not support udp now")
}
