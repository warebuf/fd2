package prediction

import (
	"fd/libs/maps"
	"fmt"
	"math"
)

func MeetsPoint(path [][]float64, pos []float64, SPD float64) float64 {

	accum := 0.0
	found := false

	for i := 0; i < len(path)-1; i++ {
		d1 := maps.EuclideanDistance(path[i][0], path[i][1], pos[0], pos[1])
		d2 := maps.EuclideanDistance(pos[0], pos[1], path[i+1][0], path[i+1][1])
		length := maps.EuclideanDistance(path[i][0], path[i][1], path[i+1][0], path[i+1][1])
		fmt.Println(i, d1, d2, length, math.Abs(d1+d2-length))
		if math.Abs(d1+d2-length) < 0.01 {
			accum += d1
			found = true
			break
		} else {
			accum += length
		}
	}
	fmt.Println(path, pos, SPD, accum, found)
	if found {
		return accum / SPD
	} else {
		panic("did not find position on path")
	}

	return -1
}
