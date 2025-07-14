package maps

import "math"

func EuclideanDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func EuclideanDistance2(a1 [2]float64, a2 [2]float64) float64 {
	dx := a2[0] - a1[0]
	dy := a2[1] - a1[1]
	return math.Sqrt(dx*dx + dy*dy)
}
