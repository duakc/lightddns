package constant

import "time"

const (
	DefaultHTTPTimeout = 15 * time.Second

	DNSTypeA    = "A"
	DNSTypeAAAA = "AAAA"
)

const (
	HTTPMaxBodySize = 10 * 1024 * 1024
)
