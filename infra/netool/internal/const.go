package internal

import "time"

var L4NetworkList = []string{"tcp", "udp", "tcp4", "tcp6", "udp4", "udp6"}

const (
	DefaultHappyEyeballFallbackDelay = 300 * time.Millisecond
	DefaultDNSTTL                    = 600
)
