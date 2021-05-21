package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip address start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4     = 1 // type is ipv4 address
	typeDm       = 3 // type is domain address
	typeIPv6     = 4 // type is ipv6 address
	typeRedirect = 9

	lenIPv4              = net.IPv4len + 2 // ipv4 + 2port
	lenIPv6              = net.IPv6len + 2 // ipv6 + 2port
	lenDmBase            = 2               // 1addrLen + 2port, plus addrLen
	AddrMask        byte = 0xf
	socksVer5            = 5
	socksCmdConnect      = 1
	socksCmdUdp          = 3
	// lenHmacSha1 = 10
)

type DebugLog bool

var (
	readTimeout time.Duration

	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")

	debug DebugLog
	// smuxConfig = smux.DefaultConfig()

)

func SetReadTimeout(c *net.Conn) {
	if readTimeout != 0 {
		(*c).SetReadDeadline(time.Now().Add(readTimeout))
	}
}

// func
func Socks5HandShake(conn *net.Conn) (err error) {
	const (
		idVer     = 0
		idNmethod = 1
	)
	// version identification and method selection message in theory can have
	// at most 256 methods, plus version and nmethod field in total 258 bytes
	// the current rfc defines only 3 authentication methods (plus 2 reserved),
	// so it won't be such long in practice
	SetReadTimeout(conn)
	buf := make([]byte, 258)
	var n int
	if n, err = io.ReadAtLeast(*conn, buf, idNmethod+1); err != nil {
		return
	}
	if buf[idVer] != socksVer5 {
		log.Println(buf)
		return errVer
	}
	nmethod := int(buf[idNmethod])
	msgLen := nmethod + 2
	if n == msgLen { // handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		if _, err = io.ReadFull(*conn, buf[n:msgLen]); err != nil {
			return
		}
	} else { // error, should not get extra data
		log.Println(buf)
		return errAuthExtraData
	}
	// send confirmation: version 5, no authentication required
	if _, err = (*conn).Write([]byte{socksVer5, 0}); err != nil {
		return err
	}
	return
}

func GetLocalRequest(conn *net.Conn) (rawaddr []byte, host string, isUdp bool, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4   = 1 // type is ipv4 address
		typeDm     = 3 // type is domain address
		typeIPv6   = 4 // type is ipv6 address
		typeChange = 5 // type is ss change config

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)

	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	var n int
	SetReadTimeout(conn)
	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(*conn, buf, idDmLen+1); err != nil {
		return
	}
	// ColorL("->", buf[:10])
	// check version and cmd
	if buf[idVer] != socksVer5 {

		err = errors.New("Sock5 error: " + string(buf[idVer]))
		return
	}
	if buf[idCmd] != socksCmdConnect && buf[idCmd] != socksCmdUdp {
		err = errCmd
		return
	}
	if buf[idCmd] == socksCmdUdp {
		isUdp = true
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	case typeChange:
		reqLen = int(buf[idDmLen]) + lenDmBase - 2
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		// ColorL("hh", host)
	default:
		err = errAddrType
		return
	}
	// ColorL("hq", buf[:10])

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(*conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		fmt.Println(n, reqLen, buf)
		err = errReqExtraData
		return
	}

	rawaddr = buf[:reqLen]

	// ColorL("hm", buf[:reqLen])

	// debug.Println("addr:", rawaddr)
	if debug {
		switch buf[idType] {
		case typeIPv4:
			host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
		case typeIPv6:
			host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
		case typeDm:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
		case typeChange:
			host = string(buf[idDm0 : idDm0+buf[idDmLen]])
			// ColorL("hm", host)

			return
		}
		port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
		host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))

	return
}

func Pipe(p1, p2 net.Conn) {
	// start tunnel & wait for tunnel termination
	// p1.SetWriteDeadline(5 * time.Second)
	// p2.SetWriteDeadline(5 * time.Second)
	streamCopy := func(dst io.Writer, src io.ReadCloser, fr, to net.Addr) {
		// startAt := time.Now()
		Copy(dst, src)
		p1.Close()
		p2.Close()
		// }()
	}
	go streamCopy(p1, p2, p2.RemoteAddr(), p1.RemoteAddr())
	streamCopy(p2, p1, p1.RemoteAddr(), p2.RemoteAddr())
	// kcpBase.aliveConn--
}

const bufSize = 4096

// const bufSize = 8192

// Memory optimized io.Copy function specified for this library
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	// fallback to standard io.CopyBuffer
	buf := make([]byte, bufSize)
	return io.CopyBuffer(dst, src, buf)
}
