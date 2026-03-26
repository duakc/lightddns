package common

import "slices"

func Zero[T any]() T {
	var v T
	return v
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
	return slices.DeleteFunc(arr, func(t T) bool {
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
