package prediction

import (
	"fd/libs/maps"
	"fd/libs/pathfinding"
)

func distanceToSegment(lineP1 []float64, lineP2 []float64, point []float64) float64 {
	ABx := lineP2[0] - lineP1[0]
	ABy := lineP2[1] - lineP1[1]
	APx := point[0] - lineP1[0]
	APy := point[1] - lineP1[1]

	// Project point P onto line AB, computing parameterized position d
	// of projection on AB
	abLenSq := ABx*ABx + ABy*ABy
	d := (APx*ABx + APy*ABy) / abLenSq

	if abLenSq == 0 {
		return maps.EuclideanDistance(point[0], point[1], lineP1[0], lineP1[1])
	}

	// Clamp d to [0,1] to stay within the segment
	if d < 0 {
		d = 0
	} else if d > 1 {
		d = 1
	}

	// Compute closest point on segment
	closestX := lineP1[0] + d*ABx
	closestY := lineP1[1] + d*ABy

	return maps.EuclideanDistance(point[0], point[1], closestX, closestY)
}

type Point struct {
	X float64
	Y float64
}

// TODO: have to create the right edges too
func IlluminatedAStar(mp *maps.MapDetails, path [][]float64, start []float64, end []float64) [][]float64 {

	newMap := new(maps.MapDetails)
	newMap.Obstacles = mp.Obstacles
	newMap.Obstacle_type = mp.Obstacle_type
	newMap.IntersectionsC = make([]maps.Coordinate, 0, len(mp.IntersectionsC)/2)
	newMap.Intersections = make([][]float64, 0, len(mp.Intersections)/2)
	newMap.Edges = make([][]int, 0, len(mp.Edges)/2)
	newMap.EdgeLengths = make([][]float64, 0, len(mp.EdgeLengths)/2)

	noDups := map[Point]struct{}{}
	for i := 0; i < len(path)-1; i++ {
		for j := 0; j < len(mp.Intersections); j++ {

			p := Point{mp.Intersections[j][0], mp.Intersections[j][1]}
			_, exists := noDups[p]
			if !exists && distanceToSegment(path[i], path[i+1], mp.Intersections[j]) <= 25.0 {

				noDups[p] = struct{}{}

				newMap.Intersections = append(newMap.Intersections, mp.Intersections[j])
				newMap.IntersectionsC = append(newMap.IntersectionsC, mp.IntersectionsC[j])
				newMap.Edges = append(newMap.Edges, mp.Edges[j])
				newMap.EdgeLengths = append(newMap.EdgeLengths, mp.EdgeLengths[j])

			}
		}
	}
	s := pathfinding.Location{X: start[0], Y: start[1]}
	e := pathfinding.Location{X: end[0], Y: end[1]}
	AStarGraph := pathfinding.AStarInit(newMap, s, e)
	bestPath := AStarGraph.ComputeShortestPath()

	return bestPath
}
