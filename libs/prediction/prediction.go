package prediction

import (
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fmt"
	"time"
)

func LastMilePrediction(mp *maps.MapDetails, path [][2]float64, distanceLeft float64, targetSPD int, hunterPosition [2]float64, hunterSPD int) [][2]float64 {

	fmt.Println("-----Called LastMilePrediction-----")

	fmt.Println("targetPath", path, distanceLeft, targetSPD)

	subpath := GetSubPath(path, distanceLeft, 0.0)

	if distanceLeft == 0 {
		subpath = [][2]float64{path[len(path)-1]} // this is a shallow copy, not a deep copy beware
	}

	fmt.Println("subPath", subpath, distanceLeft, targetSPD)

	start := time.Now()
	finalPath := findCollisionPathLinear(mp, subpath, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))
	elapsed := time.Since(start)
	fmt.Println("Execution time:", elapsed)

	fmt.Println("linear path", finalPath)

	return finalPath

}

func findCollisionPathLinear(mp *maps.MapDetails, targetPath [][2]float64, distanceStart float64, targetSPD float64, hunterPos [2]float64, hunterSPD float64) [][2]float64 {

	count := 0
	increment := 0.25 // note: if the linear search is not tight enough and does not find a collision, this algorithm will fail

	if hunterSPD == 0 {
		return [][2]float64{hunterPos}
	}

	if targetSPD == 0 || len(targetPath) == 0 {
		AStarGraph := pathfinding.AStarInit(mp, hunterPos, targetPath[0]) // this should be the target's current position
		ans := AStarGraph.ComputeShortestPath()
		return ans
	}

	targetMaxTime := distanceStart / targetSPD
	var bestPath [][2]float64

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

		var newPos [2]float64
		if t == 0 {
			newPos = targetPath[0]
		} else if t == targetMaxTime { // make sure no funny rounding problem, go to destination
			newPos = targetPath[len(targetPath)-1]
		} else {
			testPath := GetSubPath(targetPath, distanceStart, distanceLeft) // distanceStart is a bit funny to use here but it works
			newPos = testPath[len(testPath)-1]
		}

		AStarGraph := pathfinding.AStarInit(mp, hunterPos, newPos)
		bestPath = AStarGraph.ComputeShortestPath()

		count += 1
		newDist := maps.GetPathDistance(bestPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= t {
			fmt.Println("found earliest hunterTime", hunterTime, count)
			fmt.Println("earliest targetTime", t, t-increment)
			if t == 0 {
				return bestPath // already at the position ? weird
			} else {
				fmt.Println("linear count", count)
				return binary(mp, targetPath, t-increment, t, distanceStart, targetSPD, hunterPos, hunterSPD, bestPath) // check between t and t-increment (binary search) for the earliest point
			}
		}
	}

	fmt.Println("linear count", count)
	fmt.Println("no collision in time, pathing to end")
	// we did not find a collision, path to the end
	return bestPath

}

func binary(mp *maps.MapDetails, targetPath [][2]float64, low float64, high float64, distanceStart float64, targetSPD float64, hunterPos [2]float64, hunterSPD float64, topPath [][2]float64) [][2]float64 {

	count := 0
	bestPath := topPath
	epsilon := 0.001

	if low < 0 {
		low = 0
	}

	for high-low > epsilon {

		mid := (low + high) / 2

		targetTravelled := mid * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}
		testPath := GetSubPath(targetPath, distanceStart, distanceLeft)
		fmt.Println("low", low, "high", high, "mid", mid, "distanceStart", distanceStart, "distanceLeft", distanceLeft)
		newPos := testPath[len(testPath)-1]

		AStarGraph := pathfinding.AStarInit(mp, hunterPos, newPos)
		calculated := AStarGraph.ComputeShortestPath()

		count += 1
		newDist := maps.GetPathDistance(calculated)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= mid {
			high = mid
			bestPath = calculated
		} else {
			low = mid
		}

	}
	fmt.Println("binary count", count)
	return bestPath
}
