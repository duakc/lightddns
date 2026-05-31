package domains

import "strings"

// CutFromHead cut the domain from head to end
// Example:
// dn = "1.2.3.4.5"
// ans = [ "1.2.3.4.5", "2.3.4.5", "3.4.5", "4.5", "5" ]
func CutFromHead(dn string) []string {
	ans := []string{
		dn,
	}
	for i := 0; i < len(dn); i++ {
		if dn[i] == '.' && i < len(dn)-1 && len(dn[i+1:]) > 0 {
			ans = append(ans, dn[i+1:])
		}
	}
	return ans
}

// CutFromEnd cut the domain from end to head
// Example:
// dn = "1.2.3.4.5"
// ans = [ "1.2.3.4.5", "1.2.3.4", "1.2.3", "1.2", "1" ]
func CutFromEnd(dn string) []string {
	ans := []string{
		dn,
	}
	for i := len(dn) - 1; i >= 0; i-- {
		if dn[i] == '.' {
			ans = append(ans, dn[:i])
		}
	}
	return ans
}

func IsSubDomain(target string, suffix string) bool {
	lowerT := strings.ToLower(target)
	lowerA := strings.ToLower(suffix)

	return lowerT == lowerA || strings.HasSuffix(lowerT, "."+lowerA)
}

// IsDomainName , copied from golang official net lib
func IsDomainName(s string) bool {
	// The root domain name is valid. See golang.org/issue/45715.
	if s == "." {
		return true
	}

	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}
