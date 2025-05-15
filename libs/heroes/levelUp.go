package heroes

import (
	"math"
	"math/rand"
)

func LevelUp(distribution []int, points int) []int {

	results := []int{}
	for i := 0; i < len(distribution); i++ {
		results = append(results, binomialDistribution(points, float64(distribution[i])/100))
	}

	return results
}

func binomialDistribution(rolls int, successRate float64) int {
	mean := float64(rolls) * successRate
	stddev := math.Sqrt(float64(rolls) * successRate * (1 - successRate))

	// Generate normal-distributed random number (Box-Muller transform)
	u1 := rand.Float64()
	u2 := rand.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)

	// Scale and round
	val := mean + z*stddev
	val = math.Round(val)

	// Clamp between 0 and n
	if val < 0 {
		val = 0
	} else if val > float64(rolls) {
		val = float64(rolls)
	}

	return int(val)
}
