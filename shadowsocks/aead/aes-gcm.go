package aead

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAESGCMEncryptAEAD(key, salt []byte, keySize int) (cipher.AEAD, error) {
	subkey := make([]byte, keySize)
	hkdfSHA1(key, salt, hkdfInfo, subkey)

	block, err := aes.NewCipher(subkey)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}
