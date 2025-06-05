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

	ans := findCollisionPath(mp, subpath, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))
	start := time.Now()
	findCollisionPathLinear(mp, subpath, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))
	elapsed := time.Since(start)
	fmt.Println("Execution time:", elapsed)

	fmt.Println("final path", ans)

	return ans

}

func findCollisionPathLinear(mp *maps.MapDetails, targetPath [][]float64, distanceStart float64, targetSPD float64, hunterPos []float64, hunterSPD float64) [][]float64 {

	count := 0

	var bestPath [][]float64

	if len(targetPath) == 0 {
		return bestPath
	}

	if hunterSPD == 0 {
		return bestPath
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
	fmt.Println("targetMaxTime", targetMaxTime)

	targetPositions := [][]float64{}
	targetTimes := []float64{}
	hunterTimes := []float64{}
	interceptOnce := false

	for t := targetMaxTime; t >= 0; t -= 0.5 {

		targetTravelled := t * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}

		var newPos []float64
		if t == targetMaxTime { // make sure no funny rounding problem, go to destination
			newPos = targetPath[len(targetPath)-1]
		} else {
			testPath := circle.GetSubPath(targetPath, distanceStart, distanceLeft)
			newPos = testPath[len(testPath)-1]
		}
		fmt.Println("newPos", newPos, "distanceStart", distanceStart, "distanceLeft", distanceLeft)

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		bestPath = AStarGraph.ComputeShortestPath()

		count += 1
		newDist := maps.GetPathDistance(bestPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= t {
			interceptOnce = true
			hunterTimes = append(hunterTimes, hunterTime)
		} else {
			hunterTimes = append(hunterTimes, -1)
		}

		targetPositions = append(targetPositions, newPos)
		targetTimes = append(targetTimes, t)

	}
	if !interceptOnce {
		fmt.Println("does not intercept once, use the destination path")
	}

	fmt.Println("hunterTimes", hunterTimes)
	fmt.Println("targetPositions", targetPositions)
	fmt.Println("targetTimes", targetTimes)

	fmt.Println("count", count)

	return bestPath
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

	if targetSPD == 0 {
		targetCurrentPos := targetPath[0] // this should be the target's current position
		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: targetCurrentPos[0], Y: targetCurrentPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		ans := AStarGraph.ComputeShortestPath()
		return ans
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
		fmt.Println("testPath", testPath, "distanceStart", distanceStart, "distanceLeft", distanceLeft)
		newPos := testPath[len(testPath)-1]

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		bestPath = AStarGraph.ComputeShortestPath()

		if distanceStart == distanceLeft {
			fmt.Println("123", testPath)
		}

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
