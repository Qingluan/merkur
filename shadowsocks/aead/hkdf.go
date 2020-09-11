package aead

import (
	"crypto/sha1"
	"io"

	"golang.org/x/crypto/hkdf"
)

var hkdfInfo = []byte("ss-subkey")

func hkdfSHA1(secret, salt, info, subkey []byte) {
	r := hkdf.New(sha1.New, secret, salt, info)
	if _, err := io.ReadFull(r, subkey); err != nil {
		panic(err) // should never happen
	}
}
