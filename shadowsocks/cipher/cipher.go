package cipher

import (
	"net"

	"github.com/Qingluan/merkur/shadowsocks/aead"
	"github.com/Qingluan/merkur/shadowsocks/stream"
	"github.com/Qingluan/merkur/shadowsocks/util"
)

// Cipher cipher interface
type Cipher interface {
	net.Conn
	// Read(p []byte) (n int, err error)
	// Write(p []byte) (n int, err error)
	KeySize() int
	Init(key []byte, conn net.Conn)
}

// NewCipher create cipher
func NewCipher(method string) Cipher {
	switch method {
	case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm", "chacha20-ietf-poly1305":
		return aead.NewCipher(method)
	default:
		return stream.NewCipher(method)
	}
}

// NewKey create key from method/password
func NewKey(method, password string) []byte {
	cipher := NewCipher(method)

	keySize := cipher.KeySize()
	return util.KDF(password, keySize)
}
