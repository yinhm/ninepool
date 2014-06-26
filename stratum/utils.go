package stratum

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
)

func randhash() string {
	randbytes := make([]byte, 4)
	rand.Read(randbytes)

	h := sha1.New()
	h.Write(randbytes)
	return hex.EncodeToString(h.Sum(nil))
}

func DecodeNtime(ntime string) (n int64, err error) {
	return strconv.ParseInt(ntime, 16, 64)
}
