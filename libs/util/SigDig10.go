package util

import "math"

func SigDig10(num float64) float64 {
	return math.Round(10000000000*num) / 10000000000
}
