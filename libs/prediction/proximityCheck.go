package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fmt"
	"math"
)

func createPartitions(path1 [][]float64, path2 [][]float64, SPD1 float64, SPD2 float64) ([][]float64, [][]float64, []float64) {

	// assume paths are at least 2 points

	i := 0
	j := 0
	point1 := path1[0]
	point2 := path2[0]

	partition1 := [][]float64{}
	partition2 := [][]float64{}
	t := []float64{}

	for i < len(path1) && j < len(path2) {

		partition1 = append(partition1, point1)
		partition2 = append(partition2, point2)

		if i+1 < len(path1) && j+1 < len(path2) {
			dist1 := maps.EuclideanDistance(point1[0], point1[1], path1[i+1][0], path1[i+1][1])
			dist2 := maps.EuclideanDistance(point2[0], point2[1], path2[j+1][0], path2[j+1][1])
			time1 := dist1 / SPD1
			time2 := dist2 / SPD2

			if time1 == time2 {
				point1 = path1[i+1]
				point2 = path2[j+1]
				i += 1
				j += 1
				t = append(t, time1)
			} else if time1 < time2 {
				point1 = path1[i+1]
				i += 1
				newDist2 := time1 * SPD2
				subline := circle.GetSubLine(point2, path2[j+1], newDist2)
				point2 = subline[1]
				t = append(t, time1)
			} else {
				point2 = path2[j+1]
				j += 1
				newDist1 := time2 * SPD1
				subline := circle.GetSubLine(point1, path1[i+1], newDist1)
				point1 = subline[1]
				t = append(t, time2)
			}
		} else {
			i += 1
			j += 1
			break
		}
	}

	for i < len(path1) {
		dist1 := maps.EuclideanDistance(point1[0], point1[1], path1[i][0], path1[i][1])
		time1 := dist1 / SPD1
		partition1 = append(partition1, path1[i])
		partition2 = append(partition2, point2) // unit2 stays where it was
		t = append(t, time1)
		point1 = path1[i] // move point1 forward
		i += 1
	}

	for j < len(path2) {
		dist2 := maps.EuclideanDistance(point2[0], point2[1], path2[j][0], path2[j][1])
		time2 := dist2 / SPD2
		partition2 = append(partition2, path2[j])
		partition1 = append(partition1, point1) // unit1 stays where it was
		t = append(t, time2)
		point2 = path2[j] // move point1 forward
		j += 1
	}

	return partition1, partition2, t
}

func movingLineRadius(subline1 [][]float64, subline2 [][]float64, tMax float64, SPD1 float64, SPD2 float64, radius float64) float64 {

	Xo1 := subline1[0][0]
	Yo1 := subline1[0][1]
	dx1 := subline1[1][0] - subline1[0][0]
	dy1 := subline1[1][1] - subline1[0][1]
	length1 := math.Sqrt(dx1*dx1 + dy1*dy1)
	Vx1 := SPD1 * dx1 / length1
	Vy1 := SPD1 * dy1 / length1

	Xo2 := subline2[0][0]
	Yo2 := subline2[0][1]
	dx2 := subline2[1][0] - subline2[0][0]
	dy2 := subline2[1][1] - subline2[0][1]
	length2 := math.Sqrt(dx2*dx2 + dy2*dy2)
	Vx2 := SPD2 * dx2 / length2
	Vy2 := SPD2 * dy2 / length2

	A := ((Vx1 - Vx2) * (Vx1 - Vx2)) + ((Vy1 - Vy2) * (Vy1 - Vy2))
	B := (2 * (Xo1 - Xo2) * (Vx1 - Vx2)) + (2 * (Yo1 - Yo2) * (Vy1 - Vy2))
	C := ((Xo1 - Xo2) * (Xo1 - Xo2)) + ((Yo1 - Yo2) * (Yo1 - Yo2)) - (radius * radius)

	return circle.QuadraticFormula(A, B, C, tMax)
}

func ProximityCheck(path1 [][]float64, path2 [][]float64, distanceLeft1 float64, distanceLeft2 float64, SPD1 float64, SPD2 float64) {

	fmt.Println("-----Called ProximityCheck-----")

	fmt.Println("Path1: ", path1)
	fmt.Println("Path2: ", path2)

	// get subpaths TODO: this is wrong, if the unit is already at the destination, subPath is length 0, but it needs to be length 1
	subPath1 := circle.GetSubPath(path1, distanceLeft1, 0)
	subPath2 := circle.GetSubPath(path2, distanceLeft2, 0)

	fmt.Println("subPath1: ", subPath1)
	fmt.Println("subPath2: ", subPath2)

	partition1, partition2, times := createPartitions(subPath1, subPath2, SPD1, SPD2)

	fmt.Println("Partition1: ", partition1)
	fmt.Println("Partition2: ", partition2)

	totalTime := 0.0
	for i := 0; i < len(partition1)-1; i++ {

		subline1 := [][]float64{partition1[i], partition1[i+1]}
		subline2 := [][]float64{partition2[i], partition2[i+1]}
		t := movingLineRadius(subline1, subline2, times[i], SPD1, SPD2, 25)
		if t < 0 {
			totalTime += times[i]
		} else {
			totalTime += t
			fmt.Println("ProximityCheck", i, totalTime)
		}
	}

}
