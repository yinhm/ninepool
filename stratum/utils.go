package stratum

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
)

func randhash() string {
	randbytes := make([]byte, 4)
	rand.Read(randbytes)

	h := sha1.New()
	h.Write(randbytes)
	return hex.EncodeToString(h.Sum(nil))
}
