package heroes

type Hero struct {
	Bot   bool
	Name  string
	Index [3]int

	Start        [2]float64
	Position     [2]float64
	Destination  [2]float64
	Path         [][2]float64
	Thetas       []float64
	Path_Segment int // is the last node we've visited

	DestinationType  int       // 0 is camp, 1 is unit
	DestinationIndex []int     // either {i,j,k}, {i}, {}
	DynamicCircle    []float64 // a location X, Y
	LastMile         int       // 0 = closing circle, 1 = last mile, 2 = collision pathing, 3 = waiting for collision
	LastMileRadius   float64

	DistanceLeft float64
	Status       int // 0 = requires destination, 1 = in route, 2 = attacking something

	H Part
	L Part
	R Part
	B Part

	MaxHP int
	HP    int
	ATK   int
	DEF   int
	AS    int
	SPD   int
	Scale []int

	EXP int

	Battle *Camp
	TTD    float64
}
