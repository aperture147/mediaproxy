package util

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMd5String(buffer *[]byte) string {
	hash := md5.Sum(*buffer)
	return hex.EncodeToString(hash[:])
}
