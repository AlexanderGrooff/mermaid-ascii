package cmd

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func CeilDiv(x, y int) int {
	if x%y == 0 {
		return x / y
	}
	return x/y + 1
}
