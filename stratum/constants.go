package stratum

import (
	"time"
)

// Stratum error codes.
const (
	ErrorUnknown            = 20
	ErrorJobNotFound        = 21
	ErrorDuplicateShare     = 22
	ErrorLowDifficultyShare = 23
	ErrorUnauthorizedWorker = 24
	ErrorUnsubscribedWorker = 25

	ExtraNonce2Size = 2
	ExtraNonce3Size = 2 // two bytes, up to 65535 clients.
)

var errorText = map[int]string{
	ErrorUnknown:            "Unknown error",
	ErrorJobNotFound:        "Job not found",
	ErrorDuplicateShare:     "Dupliate share",
	ErrorLowDifficultyShare: "Low difficulty share",
	ErrorUnauthorizedWorker: "Unauthorized worker",
	ErrorUnsubscribedWorker: "Worker not subscribed",
}

var DefaultPoolTimeout = time.Duration(10) * time.Minute
