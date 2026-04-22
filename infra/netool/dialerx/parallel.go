package dialerx

import "time"

type happyEyeballConf struct {
	fallbackDelay time.Duration
	dialStrategy  DialStrategy
}
