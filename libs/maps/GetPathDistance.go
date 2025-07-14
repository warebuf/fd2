package maps

import "math"

func GetPathDistance(path [][2]float64) float64 {
	total := 0.0
	for i := 0; i < len(path)-1; i++ {
		dx := path[i+1][0] - path[i][0]
		dy := path[i+1][1] - path[i][1]
		total += math.Sqrt(dx*dx + dy*dy)
	}
	return total
}
