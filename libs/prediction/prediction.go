package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fmt"
	"math"
)

func LastMilePrediction(mp *maps.MapDetails, path [][]float64, distanceLeft float64, targetSPD int, hunterPosition []float64, hunterSPD int) [][]float64 {

	fmt.Println("-----Called LastMilePrediction-----")

	subpath := circle.GetSubPath(path, distanceLeft, 0.0)

	if distanceLeft == 0 {
		subpath = [][]float64{path[len(path)-1]} // this is a shallow copy, not a deep copy beware
	}

	ans := findCollisionPath(mp, subpath, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))

	fmt.Println("final path", ans)

	return ans

}

func findCollisionPath(mp *maps.MapDetails, targetPath [][]float64, distanceStart float64, targetSPD float64, hunterPos []float64, hunterSPD float64) [][]float64 {

	count := 0

	var bestPath [][]float64

	if len(targetPath) == 0 {
		return bestPath
	}

	if hunterSPD == 0 {
		return bestPath
	}

	// calculate starting range for t
	endpoint := targetPath[len(targetPath)-1]
	fmt.Println("targetPath", targetPath)
	start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
	end := pathfinding.Location{X: endpoint[0], Y: endpoint[1]}
	AStarEndpoint := pathfinding.AStarInit(mp, start, end)
	bestPath = AStarEndpoint.ComputeShortestPath()
	count += 1
	if len(targetPath) == 1 {
		return bestPath
	}

	pathDist := maps.GetPathDistance(bestPath)
	high := pathDist / float64(hunterSPD)
	low := 0.0

	targetMaxTime := distanceStart / targetSPD
	if targetMaxTime > high {
		high = targetMaxTime
	}

	for high-low > 0.001 {

		mid := (low + high) / 2

		targetTravelled := mid * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}
		testPath := circle.GetSubPath(targetPath, distanceStart, distanceLeft)
		fmt.Println("testPath", testPath)
		newPos := testPath[len(testPath)-1]

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		bestPath = AStarGraph.ComputeShortestPath()

		// test123 := IlluminatedAStar(mp, testPath, hunterPos, newPos)
		// if checkDif(bestPath, test123) {
		// 	fmt.Println("normalpath", bestPath)
		// 	fmt.Println("illuminatedPath", test123)
		// }

		count += 1
		newDist := maps.GetPathDistance(bestPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= mid {
			high = mid
		} else {
			low = mid
		}

	}

	fmt.Println("count", count)

	return bestPath
}

func checkDif(a [][]float64, b [][]float64) bool {
	if len(a) != len(b) {
		return true
	}

	for i := range a {
		if len(a[i]) != len(b[i]) {
			return true
		}
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) > 1e-9 {
				return true
			}
		}
	}

	return false
}
