package stratum

import (
	"flag"
	"time"
)

type Options struct {
	SubscribeTimeout time.Duration
}

func ParseCommandLine() (options Options, err error) {
	flag.DurationVar(&options.SubscribeTimeout, "subscribeTimeout",
		time.Duration(10)*time.Second, "Subscribe timeout")
	flag.Parse()
	return options, nil
}
