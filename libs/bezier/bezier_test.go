package bezier

import (
	"testing"
)

func Test1(t *testing.T) {

	if IntersectBezierLine([]float64{0, 0, 1, 2, 2, 2, 3, 3}, []float64{4, 0, 4, 4}) {
		t.Errorf("Expected 0")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 2, 2, 2, 3, 3}, []float64{0, 0, 3, 3}) {
		t.Errorf("Expected 1")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 2, 2, 3, 4, 3}, []float64{0, 3, 4, 3}) {
		t.Errorf("Expected 1")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 3, 3, 1, 4, 4}, []float64{2, 2, 4, 2}) {
		t.Errorf("Expected 1")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 1, 2, 2, 3, 3}, []float64{0, 0, 3, 3}) {
		t.Errorf("Expected inf, lies on line")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 2, 2, 1, 3, 3}, []float64{3, 3, 4, 4}) {
		t.Errorf("Expected 1")
	}
	if IntersectBezierLine([]float64{0, 0, 1, 1, 3, 3, 4, 4}, []float64{0, 1, 4, 5}) {
		t.Errorf("Expected 0")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 2, 3, 2, 4, 4}, []float64{2, 0, 2, 5}) {
		t.Errorf("Expected 1")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 3, 3, 1, 4, 4}, []float64{0, 2, 4, 2}) {
		t.Errorf("Expected 1")
	}
	if !IntersectBezierLine([]float64{0, 0, 3, 0, 0, 3, 3, 3}, []float64{1.5, 0, 1.5, 3}) {
		t.Errorf("Expected 1")
	}
	if IntersectBezierLine([]float64{0, 0, 1, 2, 2, 2, 3, 0}, []float64{1, 6, 0, 5}) {
		t.Errorf("Expected 0")
	}
	if !IntersectBezierLine([]float64{0, 0, 1, 2, 2, 2, 3, 0}, []float64{1, 6, 0, -5}) {
		t.Errorf("Expected 1")
	}
	if IntersectBezierLine([]float64{314.4876, 372.7915, 326.8551, 362.1908, 331.2720, 332.155477, 351.5901, 351.5901}, []float64{0, 500, 180, 420}) {
		t.Errorf("Expected 0")
	}
	if IntersectBezierLine([]float64{234.982, 176.6784, 211.1307, 166.0777, 203.180, 218.197, 231.448, 213.78}, []float64{0, 500, 90, 370}) {
		t.Errorf("Expected 0")
	}
	if IntersectBezierLine([]float64{234.982, 128.09, 234.98, 128.09, 234.98, 128.09, 234.98, 154.59}, []float64{0, 500, 128, 310}) {
		t.Errorf("Expected 0")
	}
	if IntersectBezierLine([]float64{308.303, 393.99, 308.303, 393.99, 308.303, 393.992, 351.59, 351.59}, []float64{200, 440, 181.978, 389.575}) {
		t.Errorf("Expected 0")
	}
	if !IntersectBezierLine([]float64{266.7844522968198, 175.7950530035336, 265.90106007067135, 188.16254416961135, 265.90106007067135, 196.113074204947, 266.7844522968198, 199.64664310954066}, []float64{222.61484, 159.0106007, 280, 190}) {
		t.Errorf("Expected 1")
	}
	if IntersectBezierLine([]float64{233.2155477031802, 324.20494699646645, 234.0989399293286, 311.8374558303887, 234.0989399293286, 303.886925795053, 233.2155477031802, 300.3533568904594}, []float64{283.56890459363956, 320.67137809187284, 327.73851590106005, 330.38869257950535}) {
		t.Errorf("Expected 0")
	}

}
