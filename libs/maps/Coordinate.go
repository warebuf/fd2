package maps

type Coordinate struct {
	X float64
	Y float64

	Type string // only types rn: camp, unit
	I    int
	J    int
	K    int
}
