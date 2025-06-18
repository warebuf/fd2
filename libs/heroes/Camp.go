package heroes

import "fd/libs/maps"

type Camp struct {
	MaxHP int
	HP    int
	XY    maps.Coordinate
	Units [][]int // second array is of the form i, j, k for a *hero object in TCH

}
