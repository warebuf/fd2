package prediction

import "fd/libs/heroes"

func isInCycle(state [][][]*heroes.Hero, focus *heroes.Hero) bool {

	current := [3]int{focus.Index[0], focus.Index[1], focus.Index[2]}

	visited := make(map[[3]int]bool)
	setStart := true

	for {
		if first, ok := visited[current]; ok {
			if first {
				return true
			} else {
				return false
			}
		} else {
			visited[current] = setStart
			setStart = false
		}

		next := state[current[0]][current[1]][current[2]].DynamicDestination
		if len(next) > 0 {
			current = [3]int{next[0], next[1], next[2]}
		} else {
			return false
		}
	}
}

// given a set of nodes, determine if in a cycle or strand
func getCyclesAndStrands(state [][][]*heroes.Hero, points [][]int) ([][3]int, [][3]int) {

	completed := make(map[[3]int]bool)
	for i := 0; i < len(points); i++ {
		key := [3]int{points[i][0], points[i][1], points[i][2]}
		completed[key] = false
	}

	cycle := [][3]int{}
	strand := [][3]int{}
	visited := make(map[[3]int]bool)

	for i := 0; i < len(points); i++ {

		// skip if we already visited
		current := [3]int{points[i][0], points[i][1], points[i][2]}
		if done := completed[current]; done {
			continue
		}

		cycleCheck := make(map[[3]int]bool)
		path := [][3]int{}
		interested := [][3]int{}
		for {

			if _, ok := cycleCheck[current]; ok { // if we encountered a node in our current path, we have a cycle

				// iterate through path, figure out which nodes are strands and cycles
				idx := 0
				for j := 0; j < len(path); j++ {
					if path[j] == current {
						break
					}
					if interested[idx] == path[j] {
						idx += 1
					}
				}

				for j := idx; j < len(interested); j++ {
					cycle = append(cycle, interested[idx])
				}

				for j := idx - 1; j >= 0; j-- { // append in reverse for hierarchy
					strand = append(strand, interested[idx])
				}

				break
			} else {
				cycleCheck[current] = true
				path = append(path, current)
			}

			if _, found := completed[current]; found { // we incidentally found a node we are interested in
				completed[current] = true
				interested = append(interested, current)
			}

			if _, found := visited[current]; found { // we have seen this before, we do not need to explore anymore
				for j := len(interested) - 1; j >= 0; j-- { // append in reverse for hierarchy
					strand = append(strand, interested[j])
				}
				break
			} else {
				visited[current] = true
			}

			next := state[current[0]][current[1]][current[2]].DynamicDestination
			if len(next) > 0 {
				current = [3]int{next[0], next[1], next[2]}
			} else { // if we are at an end, let's save
				for j := len(interested) - 1; j >= 0; j-- { // append in reverse for hierarchy
					strand = append(strand, interested[j])
				}
				break
			}
		}
	}

	return cycle, strand
}

// can actually run through cycle again, depending on status, may not ve an actual cycle
// a LastMile 0 means that the dependency does not update for a changing path
// meaning this is not a cycle
// should probably update getCyclesAndStrands to handle this

func calcCyclePath() {

}

// if the unit is changing destinations
// determine if the changed edge creates/is inside a cycle
// if it does, run the cycle algorithm
// update any non-cycle paths that depend on the cycle too

// if the unit is updating it's path
// determine which nodes are in a cycle
// update their paths using the cycle algorithm
// update any dependencies
