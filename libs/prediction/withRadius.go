package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fmt"
)

// if one path ends earlier than the other, we do not check if the longer path is within the radius of the endpoint
func comparePartitions(path1 [][]float64, path2 [][]float64, SPD1 float64, SPD2 float64) {

	// assume paths are at least 2 points

	i := 0
	j := 0
	point1 := path1[0]
	point2 := path2[0]

	fmt.Println(path1)
	fmt.Println(path2)

	partition1 := [][]float64{}
	partition2 := [][]float64{}

	for i < len(path1) && j < len(path2) {

		partition1 = append(partition1, point1)
		partition2 = append(partition2, point2)

		if i + 1 < len(path1) && j + 1 < len(path2){
			dist1 := maps.EuclideanDistance(point1[0], point1[1],path1[i+1][0],path1[i+1][1])
			dist2 := maps.EuclideanDistance(point2[0], point2[1],path2[j+1][0],path2[j+1][1])
			time1 := dist1 / SPD1
			time2 := dist2 / SPD2

			if time1 == time2 {
				point1 = path1[i+1]
				point2 = path2[j+1]
				i += 1
				j += 1
			} else if time1 < time2 {
				point1 = path1[i+1]
				i += 1
				newDist2 := time1 * SPD2
				subline := circle.GetSubLine(point2, path2[j+1], newDist2)
				point2 = subline[1]
			} else {
				point2 = path2[j+1]
				j += 1
				newDist1 := time2 * SPD1
				subline := circle.GetSubLine(point1, path1[i+1], newDist1)
				point1 = subline[1]
			}
		} else {
			i += 1
			j += 1
			break
		}
	}

	for i < len(path1) {
		partition1 = append(partition1, path1[i])
		i += 1
		partition2 = append(partition2, path2[ len(path2) - 1])
	}

	for j < len(path2) {
		partition2 = append(partition2, path2[j])
		j += 1
		partition1 = append(partition1, path1[ len(path1) - 1 ])
	}

	fmt.Println(partition1)
	fmt.Println(partition2)




}

func WithinRadius(path1 [][]float64, path2 [][]float64, SPD1 float64, SPD2 float64) {

	comparePartitions(path1, path2, SPD1, SPD2)

}