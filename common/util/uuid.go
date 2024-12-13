package util

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"
	"math/rand"
	"regexp"
	"strings"

	"github.com/satori/go.uuid"
)

func UUID() string {
	uid := uuid.NewV1()
	return uid.String()
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func Dmd5(str string) string {
	return strings.ToUpper(Md5(str))
}

func GetHashCode(str string, count int) int {
	v := crc32.ChecksumIEEE([]byte(str))
	if v < 0 {
		v = -v
	}
	return int(v) % count
}

var isUUID = regexp.MustCompile(`^\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`)

func IsUUID(str string) bool {
	return isUUID.Match([]byte(str))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
