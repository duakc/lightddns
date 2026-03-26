package netxx

import "strings"

func IsSubDomain(target string, suffix string) bool {
	lowerT := strings.ToLower(target)
	lowerA := strings.ToLower(suffix)

	return lowerT == lowerA || strings.HasSuffix(lowerT, "."+lowerA)
}
