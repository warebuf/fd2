package heroes

import (
	"math/rand"
	"sort"
)

func RandomBudget(budget int, stats int) []int {
	if stats <= 0 || budget <= 0 {
		return make([]int, stats)
	}

	// Generate random cut points between 0 and budget
	cuts := []int{0, budget}
	for i := 0; i < stats-1; i++ {
		cuts = append(cuts, rand.Intn(budget+1)) // [0, budget]
	}

	sort.Ints(cuts)

	// Calculate differences between adjacent cuts
	result := make([]int, stats)
	for i := 0; i < stats; i++ {
		result[i] = cuts[i+1] - cuts[i]
	}

	return result
}
