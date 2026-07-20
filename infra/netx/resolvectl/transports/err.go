package transports

import (
	"fmt"

	mDns "github.com/miekg/dns"
)

type RcodeError struct {
	Code     int
	Excepted int
}

func (e *RcodeError) Error() string {
	return fmt.Sprintf("bad rcode: %s, excepted: %s", mDns.RcodeToString[e.Code], mDns.RcodeToString[e.Excepted])
}
