package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fmt"
	"time"
)

func LastMilePrediction(mp *maps.MapDetails, path [][]float64, distanceLeft float64, targetSPD int, hunterPosition []float64, hunterSPD int) [][]float64 {

	fmt.Println("-----Called LastMilePrediction-----")

	subpath := circle.GetSubPath(path, distanceLeft, 0.0)

	if distanceLeft == 0 {
		subpath = [][]float64{path[len(path)-1]} // this is a shallow copy, not a deep copy beware
	}

	start := time.Now()
	finalPath := findCollisionPathLinear(mp, subpath, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))
	elapsed := time.Since(start)
	fmt.Println("Execution time:", elapsed)

	fmt.Println("linear path", finalPath)

	return finalPath

}

func findCollisionPathLinear(mp *maps.MapDetails, targetPath [][]float64, distanceStart float64, targetSPD float64, hunterPos []float64, hunterSPD float64) [][]float64 {

	count := 0
	increment := 0.25 // note: if the linear search is not tight enough and does not find a collision, this algorithm will fail

	if len(targetPath) == 0 {
		return [][]float64{}
	}

	if hunterSPD == 0 {
		return [][]float64{}
	}

	if targetSPD == 0 {
		targetCurrentPos := targetPath[0] // this should be the target's current position
		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: targetCurrentPos[0], Y: targetCurrentPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		ans := AStarGraph.ComputeShortestPath()
		return ans
	}

	targetMaxTime := distanceStart / targetSPD
	var bestPath [][]float64

	overshot := false
	for t := 0.0; ; t += increment {

		if t > targetMaxTime {
			if overshot {
				break
			}
			t = targetMaxTime
			overshot = true
		}

		targetTravelled := t * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}

		var newPos []float64
		if t == 0 {
			newPos = targetPath[0]
		} else if t == targetMaxTime { // make sure no funny rounding problem, go to destination
			newPos = targetPath[len(targetPath)-1]
		} else {
			testPath := circle.GetSubPath(targetPath, distanceStart, distanceLeft) // distanceStart is a bit funny to use here but it works
			newPos = testPath[len(testPath)-1]
		}

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		bestPath = AStarGraph.ComputeShortestPath()

		count += 1
		newDist := maps.GetPathDistance(bestPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= t {
			fmt.Println("found earliest hunterTime", hunterTime, count)
			if t == 0 {
				return bestPath // already at the position ? weird
			} else {
				fmt.Println("linear count", count)
				return binary(mp, targetPath, t-increment, t, distanceStart, targetSPD, hunterPos, hunterSPD) // check between t and t-increment (binary search) for the earliest point
			}
		}
	}

	fmt.Println("linear count", count)
	fmt.Println("no collision in time, pathing to end")
	// we did not find a collision, path to the end
	return bestPath

}

func binary(mp *maps.MapDetails, targetPath [][]float64, low float64, high float64, distanceStart float64, targetSPD float64, hunterPos []float64, hunterSPD float64) [][]float64 {

	count := 0
	var bestPath [][]float64
	epsilon := 0.001

	for high-low > epsilon {

		mid := (low + high) / 2

		targetTravelled := mid * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}
		testPath := circle.GetSubPath(targetPath, distanceStart, distanceLeft)
		//fmt.Println("testPath", testPath, "distanceStart", distanceStart, "distanceLeft", distanceLeft)
		fmt.Println("low", low, "high", high, "mid", mid, "distanceStart", distanceStart, "distanceLeft", distanceLeft)
		newPos := testPath[len(testPath)-1]

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		bestPath = AStarGraph.ComputeShortestPath()

		count += 1
		newDist := maps.GetPathDistance(bestPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= mid {
			high = mid
		} else {
			low = mid
		}

		hunterFinal := bestPath[len(bestPath)-1]
		distToTarget := maps.EuclideanDistance(newPos[0], newPos[1], hunterFinal[0], hunterFinal[1])

		if distToTarget > 1 {
			epsilon = epsilon / 2
		}

	}
	fmt.Println("binary count", count)
	return bestPath
}
