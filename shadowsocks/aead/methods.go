package aead

// https://shadowsocks.org/en/spec/AEAD-Ciphers.html
var cipherMethods = map[string]*cipherInfo{
	"aes-128-gcm":            &cipherInfo{16, newAESGCMEncryptAEAD, newAESGCMEncryptAEAD},
	"aes-192-gcm":            &cipherInfo{24, newAESGCMEncryptAEAD, newAESGCMEncryptAEAD},
	"aes-256-gcm":            &cipherInfo{32, newAESGCMEncryptAEAD, newAESGCMEncryptAEAD},
	"chacha20-ietf-poly1305": &cipherInfo{32, newChacha20Poly1305EncryptAEAD, newChacha20Poly1305EncryptAEAD},
}
