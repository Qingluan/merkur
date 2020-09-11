package stream

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAESCTRStream(key, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(block, iv), nil
}
