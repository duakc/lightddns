package common

func Map[S any, D any](arr []S, fn func(S) D) []D {
	retArr := make([]D, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		retArr = append(retArr, fn(arr[i]))
	}
	return retArr
}
