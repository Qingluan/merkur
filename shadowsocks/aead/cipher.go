package aead

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/Qingluan/merkur/shadowsocks/bufferpool"
)

// cipherInfo cipher definition
type cipherInfo struct {
	KeySize int

	genEncryptAEAD func(key, salt []byte, keySize int) (cipher.AEAD, error)
	genDecryptAEAD func(key, salt []byte, keySize int) (cipher.AEAD, error)
}

func (ci *cipherInfo) getSaltSize() int {
	return ci.KeySize
}

// Cipher cipher
type Cipher struct {
	net.Conn

	Method string

	Enc cipher.AEAD
	Dec cipher.AEAD

	key []byte

	buffer *bytes.Buffer

	rnonce []byte
	wnonce []byte

	Info *cipherInfo
}

// NewCipher create aead cipher
func NewCipher(method string) *Cipher {
	c := &Cipher{}
	c.Method = method

	info, ok := cipherMethods[method]
	if !ok {
		panic(fmt.Errorf("unsupported method: %s", method))
	}

	c.Info = info

	c.buffer = bytes.NewBuffer(nil)

	return c
}

// Init set key and conn
func (c *Cipher) Init(key []byte, conn net.Conn) {
	c.key = key
	c.Conn = conn
}

// KeySize return key size
func (c *Cipher) KeySize() int {
	return c.Info.KeySize
}

func (c *Cipher) getEncryptAEAD(salt []byte) (s cipher.AEAD, err error) {
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}

	s, err = c.Info.genEncryptAEAD(c.key, salt, c.Info.KeySize)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Cipher) getDecryptAEAD(salt []byte) (cipher.AEAD, error) {
	return c.Info.genDecryptAEAD(c.key, salt, c.Info.KeySize)
}

func (c *Cipher) read() error {
	// read size
	sizeBuf := bufferpool.Get(c.Dec.Overhead() + 2)
	defer bufferpool.Put(sizeBuf)

	_, err := io.ReadFull(c.Conn, sizeBuf)
	if err != nil {
		return err
	}

	ret, err := c.Dec.Open(sizeBuf[:0], c.rnonce, sizeBuf, nil)
	if err != nil {
		return err
	}
	increment(c.rnonce)
	payloadSize := int(binary.BigEndian.Uint16(ret)&0x3FFF) + c.Dec.Overhead()

	// read payload
	payloadBuf := bufferpool.Get(payloadSize)
	defer bufferpool.Put(payloadBuf)

	_, err = io.ReadFull(c.Conn, payloadBuf)
	if err != nil {
		return err
	}

	ret, err = c.Dec.Open(payloadBuf[:0], c.rnonce, payloadBuf, nil)
	if err != nil {
		return err
	}
	increment(c.rnonce)
	c.buffer.Write(ret)

	return nil
}

// Read read from client
func (c *Cipher) Read(p []byte) (n int, err error) {
	if c.Dec == nil {
		salt := bufferpool.Get(c.Info.getSaltSize())
		defer bufferpool.Put(salt)

		if _, err = io.ReadFull(c.Conn, salt); err != nil {
			return 0, err
		}

		s, err := c.getDecryptAEAD(salt)
		if err != nil {
			return 0, err
		}

		c.Dec = s

		// init read nonce
		c.rnonce = make([]byte, s.NonceSize())
	}

	if c.buffer.Len() > 0 {
		return c.buffer.Read(p)
	}

	err = c.read()
	if err != nil {
		return 0, err
	}

	return c.buffer.Read(p)
}

func (c *Cipher) encrypt(dst, src []byte) (n int) {
	size := len(src)

	binary.BigEndian.PutUint16(dst, uint16(size))

	ret := c.Enc.Seal(dst[:0], c.wnonce, dst[:2], nil)
	increment(c.wnonce)
	n = len(ret)

	ret = c.Enc.Seal(dst[n:n], c.wnonce, src, nil)
	increment(c.wnonce)
	n += len(ret)

	return n
}

// Write write to client
func (c *Cipher) Write(p []byte) (n int, err error) {
	if c.Enc == nil {
		salt := bufferpool.Get(c.Info.getSaltSize())
		defer bufferpool.Put(salt)

		c.Enc, err = c.getEncryptAEAD(salt)
		if err != nil {
			return 0, err
		}

		nw, err := c.Conn.Write(salt)
		if err != nil {
			return 0, err
		}
		if nw != len(salt) {
			return 0, err
		}

		// init write nonce
		c.wnonce = make([]byte, c.Enc.NonceSize())
	}

	size := len(p)

	buf := bufferpool.Get(c.Enc.Overhead() + 2 + size + c.Enc.Overhead())
	defer bufferpool.Put(buf)

	n = c.encrypt(buf, p)

	_, err = c.Conn.Write(buf[:n])
	if err != nil {
		return 0, err
	}

	return size, nil
}

func increment(b []byte) {
	for i := range b {
		b[i]++
		if b[i] != 0 {
			return
		}
	}
}
