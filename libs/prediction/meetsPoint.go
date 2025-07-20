package prediction

import (
	"fd/libs/maps"
	"math"
)

func MeetsPoint(path [][2]float64, pos [2]float64, SPD float64) float64 {

	accum := 0.0
	found := false

	for i := 0; i < len(path)-1; i++ {
		d1 := maps.EuclideanDistance2(path[i], pos)
		d2 := maps.EuclideanDistance2(pos, path[i+1])
		length := maps.EuclideanDistance2(path[i], path[i+1])
		if math.Abs(d1+d2-length) < 0.01 {
			accum += d1
			found = true
			break
		} else {
			accum += length
		}
	}

	if found {
		return accum / SPD
	} else {
		panic("did not find position on path")
	}

	return -1
}
