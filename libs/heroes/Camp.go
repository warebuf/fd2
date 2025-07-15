package heroes

type Camp struct {
	Index [1]int

	MaxHP    int
	HP       int
	Position [2]float64
	Units    [][3]int // second array is of the form i, j, k for a *hero object in TCH

}
