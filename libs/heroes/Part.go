package heroes

type Part struct {
	SECTION int // 0 is H, 1 is L, 2 is R, 3 is B
	SERIAL  int
	NAME    string

	HP  int
	DEF int
	DOG int

	ATK int
	ACC int
	CRT int

	CD     int
	CLU    int
	Weight int

	Use_current int
	Use_outof   int

	SPD int
}
