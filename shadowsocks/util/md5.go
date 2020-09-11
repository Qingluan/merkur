package util

import "crypto/md5"

// MD5 md5 hash
func MD5(datas ...[]byte) []byte {
	h := md5.New()

	for _, data := range datas {
		_, _ = h.Write(data)
	}
	return h.Sum(nil)
}
