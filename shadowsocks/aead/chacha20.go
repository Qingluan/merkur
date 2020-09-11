package aead

import (
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"
)

func newChacha20Poly1305EncryptAEAD(key, salt []byte, keySize int) (cipher.AEAD, error) {
	subkey := make([]byte, keySize)
	hkdfSHA1(key, salt, hkdfInfo, subkey)

	return chacha20poly1305.New(subkey)
}
