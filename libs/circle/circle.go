package circle

import (
	"fd/libs/maps"
	"fd/libs/util"
	"fmt"
	"math"
)

func GetSubLine(P1 []float64, P2 []float64, newLength float64) [][]float64 {

	line := [][]float64{}

	line = append(line, P1)

	oldLength := maps.EuclideanDistance(P1[0], P1[1], P2[0], P2[1])
	fraction := math.Round(10000000000*newLength/oldLength) / 10000000000 // 10 decimal places

	deltaX := P2[0] - P1[0]
	deltaY := P2[1] - P1[1]

	newDeltaX := deltaX * fraction
	newDeltaY := deltaY * fraction

	newX := newDeltaX + P1[0]
	newY := newDeltaY + P1[1]

	line = append(line, []float64{newX, newY})

	return line
}

// TODO!!! important, change to return at least 1 point and test!!!
func GetSubPath(path [][]float64, distanceLeftStart float64, distanceLeftEnd float64) [][]float64 {

	if distanceLeftStart <= distanceLeftEnd {
		return [][]float64{}
	}

	subPath := [][]float64{}
	end := [][]float64{}

	curLength := 0.0
	for i := len(path) - 1; i >= 1; i-- {

		curLength = curLength + maps.EuclideanDistance(path[i][0], path[i][1], path[i-1][0], path[i-1][1])

		if len(end) == 0 && curLength >= distanceLeftEnd {
			end = GetSubLine(path[i-1], path[i], curLength-distanceLeftEnd)
			subPath = append(subPath, end[1])
		}

		if curLength >= distanceLeftStart {
			start := GetSubLine(path[i-1], path[i], curLength-distanceLeftStart)
			subPath = append(subPath, start[1])
			break
		}

		if len(subPath) > 0 {
			euclid := maps.EuclideanDistance(end[1][0], end[1][1], path[i-1][0], path[i-1][1])
			if euclid < 1e-10 { // CORNER CASE: if the point we are appending is not the same as the end point we appended
			} else {
				subPath = append(subPath, path[i-1])
			}
		}

	}

	reversedSubPath := [][]float64{}
	for i := len(subPath) - 1; i >= 0; i-- {
		reversedSubPath = append(reversedSubPath, subPath[i])
	}

	return reversedSubPath

}

func lineIntersectsCircle(line [][]float64, cX float64, cY float64, r0 float64, r1 float64) float64 {

	p0x := line[0][0]
	p0y := line[0][1]
	p1x := line[1][0]
	p1y := line[1][1]

	t2_x := (p0x * p0x) + (p1x * p1x) - (2 * p0x * p1x)
	t_x := (-2 * p0x * p0x) + (2 * p0x * p1x) - (2 * p1x * cX) + (2 * p0x * cX)
	x := (p0x * p0x) - (2 * p0x * cX) + (cX * cX)

	t2_y := (p0y * p0y) + (p1y * p1y) - (2 * p0y * p1y)
	t_y := (-2 * p0y * p0y) + (2 * p0y * p1y) - (2 * p1y * cY) + (2 * p0y * cY)
	y := (p0y * p0y) - (2 * p0y * cY) + (cY * cY)

	t2_r := (r1 * r1) - (2 * r0 * r1) + (r0 * r0)
	t_r := (-2 * r0 * r0) + (2 * r0 * r1)
	r := r0 * r0

	A := t2_x + t2_y - t2_r
	B := t_x + t_y - t_r
	C := x + y - r

	intersect := QuadraticFormula(A, B, C, 1)

	return intersect
}

func QuadraticFormula(A float64, B float64, C float64, tMax float64) float64 { // returns the first t between 0 and 1, returns -1 if no roots

	if util.SigDig10(A) == 0 && util.SigDig10(B) == 0 {
		if C == 0 {
			return 0
		} else {
			return -1
		}
	} else if util.SigDig10(A) == 0 {
		t := -C / B
		if t >= 0 && t <= tMax {
			return t
		} else {
			return -1
		}
	}

	discriminant := (B * B) - (4 * A * C)
	if discriminant < 0 {
		return -1 // no roots
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-B - sqrtDisc) / (2 * A)
	t2 := (-B + sqrtDisc) / (2 * A)

	if t1 >= 0 && t1 <= tMax {
		return t1
	}
	if t2 >= 0 && t2 <= tMax {
		return t2
	}

	return -1
}

func CircleIntersectPath(path [][]float64, targetX float64, targetY float64, distanceLeftStart float64, distanceLeftEnd float64, circle []float64, radiusStart float64, radiusEnd float64, hunterSPD int, reductionFactor float64) float64 {

	fmt.Println("-----Called CircleIntersectPath-----")

	radiusStart = radiusStart / reductionFactor
	radiusEnd = radiusEnd / reductionFactor

	if distanceLeftEnd < 0 {
		distanceLeftEnd = 0
	}

	if distanceLeftStart == distanceLeftEnd {
		euclid := maps.EuclideanDistance(targetX, targetY, circle[0], circle[1])
		if euclid <= radiusStart && euclid >= radiusEnd {
			startToEuclid := radiusStart - euclid
			return startToEuclid * reductionFactor / float64(hunterSPD)
		} else {
			return -1
		}
	}

	subPath := GetSubPath(path, distanceLeftStart, distanceLeftEnd)

	// create an array of floats that determine the percentage each line takes of the total path
	distances := []float64{}
	total := 0.0
	for i := 0; i < len(subPath)-1; i++ {
		euclid := maps.EuclideanDistance(subPath[i][0], subPath[i][1], subPath[i+1][0], subPath[i+1][1])
		distances = append(distances, euclid)
		total += euclid
	}

	percentage := []float64{}
	for i := 0; i < len(distances); i++ {
		percentage = append(percentage, math.Round(10000000000*distances[i]/total)/10000000000)
	}

	radius := []float64{radiusStart}
	radiusChange := radiusStart - radiusEnd
	percentageCount := 0.0
	for i := 0; i < len(percentage); i++ {
		percentageCount += percentage[i]
		radius = append(radius, radiusStart-(percentageCount*radiusChange))
	}

	ans := -1.0
	percentageCount = 0.0
	intersectFound := false

	for i := 0; i < len(subPath)-1; i++ {
		line := [][]float64{subPath[i], subPath[i+1]}
		intersects := lineIntersectsCircle(line, circle[0], circle[1], radius[i], radius[i+1])

		if intersects >= 0 {
			percentageCount += percentage[i] * intersects
			intersectFound = true
			break
		}

		percentageCount += percentage[i]
	}

	if intersectFound {
		ans = (radiusChange * percentageCount * reductionFactor) / float64(hunterSPD)
	}

	return ans
}
