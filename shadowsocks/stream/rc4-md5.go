package stream

import (
	"crypto/cipher"
	"crypto/rc4"

	"github.com/Qingluan/merkur/shadowsocks/util"
)

func newRC4MD5Stream(key, iv []byte) (cipher.Stream, error) {
	rc4key := util.MD5(key, iv)

	return rc4.NewCipher(rc4key)
}
