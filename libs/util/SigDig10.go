package util

import "math"

func SigDig10(num float64) float64 {
	return math.Round(10000000000*num) / 10000000000
}

func SigDig5(num float64) float64 {
	return math.Round(100000*num) / 100000
}

func SigDig3(num float64) float64 {
	return math.Round(1000*num) / 1000
}
