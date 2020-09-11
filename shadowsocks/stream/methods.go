package stream

var cipherMethods = map[string]*cipherInfo{
	"aes-128-cfb":   &cipherInfo{16, 16, newAESCFBEncryptStream, newAESCFBDecryptStream},
	"aes-192-cfb":   &cipherInfo{24, 16, newAESCFBEncryptStream, newAESCFBDecryptStream},
	"aes-256-cfb":   &cipherInfo{32, 16, newAESCFBEncryptStream, newAESCFBDecryptStream},
	"aes-128-ctr":   &cipherInfo{16, 16, newAESCTRStream, newAESCTRStream},
	"aes-192-ctr":   &cipherInfo{24, 16, newAESCTRStream, newAESCTRStream},
	"aes-256-ctr":   &cipherInfo{32, 16, newAESCTRStream, newAESCTRStream},
	"des-cfb":       &cipherInfo{8, 8, newDESCFBEncryptStream, newDESCFBDecryptStream},
	"rc4-md5":       &cipherInfo{16, 16, newRC4MD5Stream, newRC4MD5Stream},
	"rc4-md5-6":     &cipherInfo{16, 6, newRC4MD5Stream, newRC4MD5Stream},
	"chacha20":      &cipherInfo{32, 8, newChaCha20Stream, newChaCha20Stream},
	"chacha20-ietf": &cipherInfo{32, 12, newChaCha20Stream, newChaCha20Stream},
	"cast5-cfb":     &cipherInfo{16, 8, newCast5EncryptStream, newCast5DecryptStream},
	"bf-cfb":        &cipherInfo{16, 8, newBlowfishEncryptStream, newBlowfishDecryptStream},
}
