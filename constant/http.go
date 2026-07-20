package constant

import (
	"fmt"
	"runtime"
	"time"
)

const (
	DefaultHTTPTimeout = 15 * time.Second
)

const (
	HTTPMaxBodySize = 10 * 1024 * 1024
)

var HTTPUserAgent = fmt.Sprintf("%s/%s (Golang/%s; %s/%s;)",
	Project, Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
