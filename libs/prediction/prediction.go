package prediction

import (
	"fd/libs/circle"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fmt"
)

func LastMilePrediction(mp *maps.MapDetails, path [][]float64, distanceLeft float64, targetSPD int, hunterPosition []float64, hunterSPD int) {

	subpath := circle.GetSubPath(path, distanceLeft, 0.0);

	if distanceLeft == 0 {
		subpath = [][]float64{ path[len(path)-1] } // this is a shallow copy, not a deep copy beware
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

}