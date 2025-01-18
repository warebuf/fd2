package bezier

import "fmt"
import "math"

func IntersectBezierLine(bezier []float64, line []float64) bool {

	fmt.Println("v02", bezier, line)

	// Compute line coefficients A, B, C
	A := line[1] - line[3]                         // Y1-Y2
	B := line[2] - line[0]                         // X2-X1
	C := (line[3] * line[0]) - (line[2] * line[1]) // Y2*X1 - X2*Y1

	fmt.Println("A:", A, "B:", B, "C:", C)

	// Compute Bézier coefficients
	x0, y0 := bezier[0], bezier[1]
	x1, y1 := bezier[2], bezier[3]
	x2, y2 := bezier[4], bezier[5]
	x3, y3 := bezier[6], bezier[7]

	a := A*(x3-3*x2+3*x1-x0) + B*(y3-3*y2+3*y1-y0)
	b := A*(3*x2-6*x1+3*x0) + B*(3*y2-6*y1+3*y0)
	c := A*(3*x1-3*x0) + B*(3*y1-3*y0)
	d := C

	fmt.Println("abcd", a, b, c, d)

	roots := cardano(a, b, c, d)

	// Check if any root is valid and lies within the line segment
	for _, t := range roots {
		if t >= 0 && t <= 1 { // Root is within Bézier parameter range
			XY := calcXY(bezier, t)
			if XY[0] >= math.Min(line[0], line[2]) && XY[0] <= math.Max(line[0], line[2]) &&
				XY[1] >= math.Min(line[1], line[3]) && XY[1] <= math.Max(line[1], line[3]) {
				return true // Intersection found
			}
		}
	}

	return false
}

func calcXY(bezier []float64, t float64) []float64 {

	X1 := (1 - t) * (1 - t) * (1 - t) * bezier[0]
	X2 := (1 - t) * (1 - t) * (t) * bezier[2] * 3
	X3 := (1 - t) * (t) * (t) * bezier[4] * 3
	X4 := (t) * (t) * (t) * bezier[6]

	X := X1 + X2 + X3 + X4

	Y1 := (1 - t) * (1 - t) * (1 - t) * bezier[1]
	Y2 := (1 - t) * (1 - t) * (t) * bezier[3] * 3
	Y3 := (1 - t) * (t) * (t) * bezier[5] * 3
	Y4 := (t) * (t) * (t) * bezier[7]

	Y := Y1 + Y2 + Y3 + Y4

	fmt.Println("X:", X, "Y:", Y)

	return []float64{X, Y}
}

func cardano(a, b, c, d float64) []float64 {

	// Normalize coefficients
	if a != 0 {
		b /= a
		c /= a
		d /= a
	}

	// Convert to depressed cubic: t^3 + pt + q = 0
	p := c - b*b/3
	q := 2*b*b*b/27 - b*c/3 + d
	discriminant := q*q/4 + p*p*p/27

	var roots []float64

	if discriminant > 0 {
		// One real root and two complex roots
		sqrtDisc := math.Sqrt(discriminant)
		u := math.Cbrt(-q/2 + sqrtDisc)
		v := math.Cbrt(-q/2 - sqrtDisc)
		root := u + v - b/3
		roots = append(roots, root)
	} else if discriminant == 0 {
		// All roots real, at least two equal
		if q == 0 {
			root := -b / 3
			roots = append(roots, root, root, root)
		} else {
			u := math.Cbrt(-q / 2)
			root1 := 2*u - b/3
			root2 := -u - b/3
			roots = append(roots, root1, root2)
		}
	} else {
		// Three distinct real roots
		r := math.Sqrt(-p * p * p / 27)
		phi := math.Acos(-q / (2 * r))
		rCbrt := math.Cbrt(r)

		root1 := 2*rCbrt*math.Cos(phi/3) - b/3
		root2 := 2*rCbrt*math.Cos((phi+2*math.Pi)/3) - b/3
		root3 := 2*rCbrt*math.Cos((phi+4*math.Pi)/3) - b/3

		roots = append(roots, root1, root2, root3)
	}

	return roots
}
