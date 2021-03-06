package shopee

type number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}

func min[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}
