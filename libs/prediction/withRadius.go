package prediction

import (
	"fd/libs/maps"
)

// if one path ends earlier than the other, we do not check if the longer path is within the radius of the endpoint

func withinRadius(path1 [][]float64, path2 [][]float64, SPD1 float64, SPD2 float64) {

	// assume paths are at least 2 points

	i := 0
	j := 0
	point1 := path1[0]
	point2 := path2[0]

	partition1 := [][]float64{}
	partition2 := [][]float64{}

	for len(point1) > 0  && len(point2) > 0 {

		partition1 = append(partition1, point1)
		partition2 = append(partition2, point2)

		dist1 := maps.EuclideanDistance(point1[0], point1[1],path1[i+1][0],path1[i+1][1])
		dist2 := maps.EuclideanDistance(point2[0], point2[1],path2[j+1][0],path2[j+1][1])
		time1 := dist1 / SPD1
		time2 := dist2 / SPD2

		if time1 == time2 {
			point1 = path1[i]
			i += 1
			point2 = path2[j]
			j += 1

		} else if time1 < time2 {
			i += 1
			point1 = path1[i]


		} else {

		}

	}





}