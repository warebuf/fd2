package bezier

import (
	"math"
)

func IntersectBezierLine(bezier []float64, line []float64) bool {

	//fmt.Println("v12", bezier, line)

	// Compute line coefficients A, B, C
	A := line[1] - line[3]                         // Y1-Y2(X)
	B := line[2] - line[0]                         // X2-X1(Y)
	C := (line[3] * line[0]) - (line[2] * line[1]) // Y2*X1 - X2*Y1

	//fmt.Println("A:", A, "B:", B, "C:", C) // THIS IS CORRECT, CALCULATE IT MANUALLY Y=MX+B

	// THE PROBLEM LIES SOMEWHERE HERE CALCULATING THE COEFFICIENT
	// Compute Bézier coefficients
	x0, y0 := bezier[0], bezier[1]
	x1, y1 := bezier[2], bezier[3]
	x2, y2 := bezier[4], bezier[5]
	x3, y3 := bezier[6], bezier[7]

	a := A*(x3-3*x2+3*x1-x0) + B*(y3-3*y2+3*y1-y0)
	b := A*(3*x2-6*x1+3*x0) + B*(3*y2-6*y1+3*y0)
	c := A*(3*x1-3*x0) + B*(3*y1-3*y0)
	d := (A * x0) + (B * y0) + C

	//fmt.Println("abcd", a, b, c, d)

	// if a == 0, corner case, checkQuadraticRoots
	EPSILON := 0.0000000001
	if math.Abs(a) < EPSILON {
		if math.Abs(b) < EPSILON {
			return checkLinearRoots(c, d, bezier, line)
		}
		return checkQuadraticRoots(b, c, d, bezier, line)
	}

	// CARDANOS IS CORRECT, VERIFY BY ONLINE CUBIC EQUATION SOLVER
	roots := cardano(a, b, c, d)
	//fmt.Println("cardano", roots)

	// Check if any root is valid and lies within the line segment
	for _, t := range roots {
		if t >= 0 && t <= 1 { // Root is within Bézier parameter range
			XY := calcXY(bezier, t)
			if onLineSegment(XY, line) {
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

	//fmt.Println("X:", X, "Y:", Y)

	return []float64{X, Y}
}

func cardano(a, b, c, d float64) []float64 {

	// Normalize coefficients
	if a != 0 {
		b /= a
		c /= a
		d /= a
	}

	//fmt.Println("normalized", "b:", b, "c:", c, "d:", d)

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

// bt^2 + ct + d = 0
func checkQuadraticRoots(a float64, b float64, c float64, bezier []float64, line []float64) bool {

	discriminant := (b * b) - (4 * a * c)
	if discriminant < 0 {
		return false
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-b + sqrtDisc) / (2 * a)
	t2 := (-b - sqrtDisc) / (2 * a)

	if t1 >= 0 && t1 <= 1 {
		XY := calcXY(bezier, t1)
		if onLineSegment(XY, line) {
			return true // Intersection found
		}
	}

	if t2 >= 0 && t2 <= 1 {
		XY := calcXY(bezier, t2)
		if onLineSegment(XY, line) {
			return true // Intersection found
		}
	}

	return false
}

func checkLinearRoots(a float64, b float64, bezier []float64, line []float64) bool {

	if a == 0 {
		if b == 0 {
			return true
		}
		return false
	}

	root := -b / a

	if root >= 0 && root <= 1 {
		XY := calcXY(bezier, root)
		if onLineSegment(XY, line) {
			return true // Intersection found
		}
	}

	return false
}

// if X is within 6 sig digs or greater
func onLineSegment(point []float64, line []float64) bool {
	if point[0] < math.Min(line[0], line[2])-0.000001 {
		return false
	}
	if point[0] > math.Max(line[0], line[2])+0.000001 {
		return false
	}
	if point[1] < math.Min(line[1], line[3])-0.000001 {
		return false
	}
	if point[1] > math.Max(line[1], line[3])+0.000001 {
		return false
	}
	return true
}
