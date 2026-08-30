// Package numcompare provides numeric comparison helpers shared by provider models.
package numcompare

import "cmp"

// TTL compares DNS TTL values. A zero TTL means that the provider should use
// its default value; because APIs return that resolved value, zero is treated
// as equal to any non-zero TTL to avoid an update on every reconciliation.
func TTL(a, b uint32) int {
	if a == 0 || b == 0 {
		return 0
	}
	return cmp.Compare(a, b)
}

// Bool compares booleans using the same ordering as other comparison helpers.
func Bool(a, b bool) int {
	if a == b {
		return 0
	}
	if a {
		return 1
	}
	return -1
}
