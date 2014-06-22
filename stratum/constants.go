package stratum

// Stratum error codes.
const (
	ErrorUnauthorizedWorker = 24
	ErrorUnsubscribedWorker = 25
)

var errorText = map[int]string{
	ErrorUnauthorizedWorker: "Unauthorized worker",
	ErrorUnsubscribedWorker: "Worker not subscribed",
}
