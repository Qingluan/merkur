package stream

import (
	"crypto/cipher"

	"github.com/aead/chacha20"
)

func newChaCha20Stream(key, iv []byte) (cipher.Stream, error) {
	return chacha20.NewCipher(iv, key)
}
