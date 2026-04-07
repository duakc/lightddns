package common

import (
	"cmp"
	"maps"
	"slices"
	"strings"
)

func Zero[T any]() T {
	var v T
	return v
}

func Comparable[T comparable](v T) T {
	return v
}

func PtrValueOrDefault[T any](v *T) T {
	if v == nil {
		return Zero[T]()
	}
	return *v
}

func All[T any, S ~[]T](arr S, fn func(T) bool) bool {
	for i := 0; i < len(arr); i++ {
		if !fn(arr[i]) {
			return false
		}
	}
	return true
}

func Or[T any, S ~[]T](arr S, fn func(T) bool) bool {
	for i := 0; i < len(arr); i++ {
		if fn(arr[i]) {
			return true
		}
	}
	return false
}

func Filter[T any, S ~[]T](arr S, fn func(T) bool) S {
	return slices.DeleteFunc(slices.Clone(arr), func(t T) bool {
		return !fn(t)
	})
}

func ToMap[T comparable, S ~[]T](arr S) map[T]bool {
	m := make(map[T]bool)
	for i := 0; i < len(arr); i++ {
		m[arr[i]] = true
	}
	return m
}

func Reduce[T any, S ~[]T](arr S, fn func(v1, v2 T) T) T {
	if len(arr) == 0 {
		return Zero[T]()
	}
	piovt := arr[0]
	for i := 1; i < len(arr); i++ {
		piovt = fn(piovt, arr[i])
	}
	return piovt
}

func Sum[T cmp.Ordered, S ~[]T](arr S) T {
	var su T
	for i := 0; i < len(arr); i++ {
		su += arr[i]
	}
	return su
}

func MergeMap[K comparable, V any](mm ...map[K]V) map[K]V {
	allLen := Sum(Map(mm, func(s map[K]V) int {
		return len(mm)
	}))
	res := make(map[K]V, allLen)
	for _, m := range mm {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

func UnquoteString(s string) string {
	trim := strings.Trim(s, "'\"\n\r\t")
	return strings.TrimSpace(trim)
}

func Distinct[T comparable, S ~[]T](arr S) S {
	return slices.Collect(maps.Keys(ToMap(arr)))
}
