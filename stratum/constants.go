package stratum

// Stratum error codes.
const (
	ErrorUnknown            = 20
	ErrorJobNotFound        = 21
	ErrorDuplicateShare     = 22
	ErrorLowDifficultyShare = 23
	ErrorUnauthorizedWorker = 24
	ErrorUnsubscribedWorker = 25
)

var errorText = map[int]string{
	ErrorUnknown:            "Unknown error",
	ErrorJobNotFound:        "Job not found",
	ErrorDuplicateShare:     "Dupliate share",
	ErrorLowDifficultyShare: "Low difficulty share",
	ErrorUnauthorizedWorker: "Unauthorized worker",
	ErrorUnsubscribedWorker: "Worker not subscribed",
}
