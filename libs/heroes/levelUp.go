package heroes

import (
	"math"
	"math/rand"
)

func LevelUp(unit *Hero, rolls int) {

	statChanges := LevelUpRoll(unit.Scale, rolls)

	unit.MaxHP += statChanges[0]
	unit.HP += statChanges[0]
	unit.ATK += statChanges[1]
	unit.AS += statChanges[2]
	unit.DEF += statChanges[3]
	unit.SPD += statChanges[4]

}

// given probabilities (1-100) and number of chances, generate an array of how much each stat increases for a level up
func LevelUpRoll(distribution []int, rolls int) []int {

	results := []int{}
	for i := 0; i < len(distribution); i++ {
		results = append(results, binomialDistribution(rolls, float64(distribution[i])/100))
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
