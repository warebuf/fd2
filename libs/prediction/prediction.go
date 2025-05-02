package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fmt"
)

func LastMilePrediction(mp *maps.MapDetails, path [][]float64, distanceLeft float64, targetSPD int, hunterPosition []float64, hunterSPD int) {

	fmt.Println("-----Called LastMilePrediction-----")

	subpath := circle.GetSubPath(path, distanceLeft, 0.0)

	if distanceLeft == 0 {
		subpath = [][]float64{path[len(path)-1]} // this is a shallow copy, not a deep copy beware
	}

	fmt.Println("subpath", subpath)

	possibleInterceptCoordinates := [][]float64{}
	possibleInterceptTimes := []float64{}

	constructingPath := [][]float64{}
	for i := 0; i < len(subpath); i++ {
		fmt.Println("subpath[i]", subpath[i])
		constructingPath = append(constructingPath, subpath[i])

		// compute A* from current position to subpoint
		start := pathfinding.Location{X: hunterPosition[0], Y: hunterPosition[1]}
		end := pathfinding.Location{X: subpath[i][0], Y: subpath[i][1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		aStarPath := AStarGraph.ComputeShortestPath()
		fmt.Println("aStarPath", aStarPath)
		hunterDistance := maps.GetPathDistance(aStarPath)
		fmt.Println("hunterDistance", hunterDistance)
		targetDistance := maps.GetPathDistance(constructingPath)
		fmt.Println("targetDistance", targetDistance)
		hunterTime := hunterDistance / float64(hunterSPD)
		fmt.Println("hunterTime", hunterTime)
		targetTime := targetDistance / float64(targetSPD)
		fmt.Println("targetTime", targetTime)
		if hunterTime <= targetTime {
			fmt.Println("found earlier time", "hunterTime", hunterTime, "targetTime", targetTime)
			possibleInterceptCoordinates = append(possibleInterceptCoordinates, subpath[i])
			possibleInterceptTimes = append(possibleInterceptTimes, hunterTime)
		}
	}

	fmt.Println("possibleInterceptCoordinates", possibleInterceptCoordinates)
	fmt.Println("possibleInterceptTimes", possibleInterceptTimes)

	// if no intercepts, just direct the unit towards the end of the path
	if len(possibleInterceptTimes) == 0 {
		fmt.Println("no intercepts, direct the unit to the end of the path")
	}

	ans := earliest(mp, path, distanceLeft, float64(targetSPD), hunterPosition, float64(hunterSPD))

	fmt.Println("ans", ans)
}

func earliest(mp *maps.MapDetails, targetPath [][]float64, distanceStart float64, targetSPD float64, hunterPos []float64, hunterSPD float64) float64 {

	if len(targetPath) == 0 {
		return 0
	}

	if hunterSPD == 0 {
		return 0
	}

	// calculate starting range for t
	endpoint := targetPath[len(targetPath)-1]
	start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
	end := pathfinding.Location{X: endpoint[0], Y: endpoint[1]}
	AStarEndpoint := pathfinding.AStarInit(mp, start, end)
	path := AStarEndpoint.ComputeShortestPath()
	pathDist := maps.GetPathDistance(path)
	high := pathDist / float64(hunterSPD)
	low := 0.0

	for high-low > 0.01 {

		mid := (low + high) / 2

		targetTravelled := mid * targetSPD
		distanceLeft := distanceStart - targetTravelled
		if distanceLeft < 0 {
			distanceLeft = 0
		}
		testPath := circle.GetSubPath(targetPath, distanceStart, distanceLeft)
		newPos := testPath[len(testPath)-1]

		start := pathfinding.Location{X: hunterPos[0], Y: hunterPos[1]}
		end := pathfinding.Location{X: newPos[0], Y: newPos[1]}
		AStarGraph := pathfinding.AStarInit(mp, start, end)
		newPath := AStarGraph.ComputeShortestPath()
		newDist := maps.GetPathDistance(newPath)
		hunterTime := newDist / float64(hunterSPD)

		if hunterTime <= mid {
			high = mid
		} else {
			low = mid
		}

	}

	return high
}

//function get_position_along_path(path, speed, time):
//distance = speed * time
//for i from 0 to len(path)-2:
//segment = distance_between(path[i], path[i+1])
//if distance < segment:
//return interpolate(path[i], path[i+1], distance / segment)
//distance -= segment
//return path[-1]  # reached end
