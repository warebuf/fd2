package maps

type MapDetails struct {
	Obstacles     [][]float64 // [x1,y1,x2,y2,x3,y3,x4,y4]
	Obstacle_type []string    // [CIRCLE,BEZIER,...]

	IntersectionsC           []Coordinate
	Intersections            [][]float64 // [x,y]
	Intersections_edges_int  [][]int
	Intersections_edgelength [][]float64 // [el1, el2, ...]
}
