package prediction

import (
	"fmt"
	"math"
)

func GainRadius(hunterStart, hunterEnd, targetStart, targetEnd []float64, hunterSPD, targetSPD int) bool {
	fmt.Println("-----Called GainRadius-----")
	fmt.Println(hunterStart, hunterEnd, targetStart, targetEnd)
	fmt.Println(hunterSPD, targetSPD)

	// Compute direction vectors
	hx := hunterEnd[0] - hunterStart[0]
	hy := hunterEnd[1] - hunterStart[1]
	hl := math.Hypot(hx, hy)

	tx := targetEnd[0] - targetStart[0]
	ty := targetEnd[1] - targetStart[1]
	tl := math.Hypot(tx, ty)

	vHunter := []float64{0, 0}
	vTarget := []float64{0, 0}

	if hl != 0 {
		vHunter[0] = float64(hunterSPD) * hx / hl
		vHunter[1] = float64(hunterSPD) * hy / hl
	}
	if tl != 0 {
		vTarget[0] = float64(targetSPD) * tx / tl
		vTarget[1] = float64(targetSPD) * ty / tl
	}

	// Relative position and velocity
	pRel := []float64{hunterStart[0] - targetStart[0], hunterStart[1] - targetStart[1]}
	vRel := []float64{vHunter[0] - vTarget[0], vHunter[1] - vTarget[1]}

	// Dot product
	dot := pRel[0]*vRel[0] + pRel[1]*vRel[1]
	fmt.Println("dot", dot)

	return dot > 0 // true = hunter is losing ground
}
