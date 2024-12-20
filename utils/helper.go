package utils

import "strconv"

func StringToFloat64(s string) float64 {
	result, _ := strconv.ParseFloat(s, 64)

	return result
}
