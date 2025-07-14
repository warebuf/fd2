package maps

type MapDetails struct {
	Obstacles     [][]float64 // [x1,y1,x2,y2,x3,y3,x4,y4]
	Obstacle_type []string    // [CIRCLE,BEZIER,...]

	Intersections [][2]float64
	Edges         [][]int
	EdgeLengths   [][]float64
}
