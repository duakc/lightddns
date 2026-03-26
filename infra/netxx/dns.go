package netxx

import "errors"

var (
	ErrNoDNSRecord = errors.New("no dns record found")
)

type DNSRecord struct {
	Type    string // only support `A` and `AAAA` now
	Name    string
	Content string

	TTL int
}
