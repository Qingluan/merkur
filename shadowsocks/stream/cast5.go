package stream

import (
	"crypto/cipher"

	"golang.org/x/crypto/cast5"
)

func newCast5EncryptStream(key, iv []byte) (cipher.Stream, error) {
	block, err := cast5.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewCFBEncrypter(block, iv), nil
}

func newCast5DecryptStream(key, iv []byte) (cipher.Stream, error) {
	block, err := cast5.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewCFBDecrypter(block, iv), nil
}
