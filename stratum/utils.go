package stratum

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"github.com/conformal/fastsha256"
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

func DoubleSha256(b []byte) []byte {
	hasher := fastsha256.New()
	hasher.Write(b)
	sum := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write(sum)
	return hasher.Sum(nil)
}

func CoinbaseHash(coinbase1, nonce1, nonce2, coinbase2 string) string {
	coinbase := coinbase1 + nonce1 + nonce2 + coinbase2
  buf, _ := hex.DecodeString(coinbase)
	return hex.EncodeToString(DoubleSha256(buf))
}
