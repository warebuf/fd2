package pathfinding

import (
	"fmt"
	"math"

	"fd/libs/maps"
)

type lpaNode struct {
	X float64
	Y float64

	G   float64 // Cost from start to this node
	H   float64
	RHS float64

	Succ *[]lpaNode
	Pred *[]lpaNode
}

type LPAUnitMemory struct {

	// PQ
	LPANodes []*lpaNode
}

func euclideanDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func calcLPAKey(s *lpaNode, goal *lpaNode) (float64, float64) {
	var newG, newRHS float64
	euclid := euclideanDistance(s.X, s.Y, goal.X, goal.Y)

	if s.G < s.RHS+euclid {
		newG = s.G
	} else {
		newG = s.RHS + euclid
	}

	if s.G < s.RHS {
		newRHS = s.G
	} else {
		newRHS = s.RHS
	}

	return newG, newRHS
}

func initLPA(m *maps.MapDetails) {

	fmt.Println("len(m.Intersections)", len(m.Intersections))
	fmt.Println("len(m.Intersections_edges_int)", len(m.Intersections_edges_int))
	fmt.Println("len(m.Intersections_edgelength)", len(m.Intersections_edgelength))

	for i := 0; i < len(m.Intersections); i++ {

		fmt.Println(m.Intersections[i])
		fmt.Println(m.Intersections_edges_int[i])
		fmt.Println(m.Intersections_edgelength[i])
		
	}

}

func LPA(m *maps.MapDetails) {
	fmt.Println("calling LPA*")
	initLPA(m)

}
