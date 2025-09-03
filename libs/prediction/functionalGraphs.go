package prediction

import (
	"fd/libs/heroes"
)

// given a set of nodes, determine if in a cycle or strand
func GetCyclesAndStrands(state [][][]*heroes.Hero, points [][3]int, hardUpdate map[[3]int]int) ([][3]int, [][3]int) {

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

		if state[current[0]][current[1]][current[2]].HP <= 0 {
			visited[current] = true
			completed[current] = true
			strand = append(strand, current)
			continue
		}

		cycleCheck := make(map[[3]int]bool)
		path := [][3]int{}       // the nodes that we have visited
		interested := [][3]int{} // the nodes that we have visited that are points
		for {

			if _, ok := cycleCheck[current]; ok { // if we encountered a node in our current path, we have a cycle

				// iterate through path, figure out which nodes are strands and cycles
				idx := 0
				for j := 0; j < len(path); j++ {
					if path[j] == current {
						break
					}
					if idx < len(interested) && interested[idx] == path[j] {
						idx += 1
					}
				}

				for j := idx; j < len(interested); j++ {
					cycle = append(cycle, interested[j])
				}

				for j := idx - 1; j >= 0; j-- { // append in reverse for hierarchy
					strand = append(strand, interested[j])
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

			next := state[current[0]][current[1]][current[2]].DestinationIndex
			status, ok := hardUpdate[current]
			stayLastMile0 := !ok && state[current[0]][current[1]][current[2]].LastMile == 0
			changeToLastMile0 := ok && status == 0                                                                // lm0 will never be an effective cycle
			if stayLastMile0 || changeToLastMile0 || len(next) <= 1 || state[next[0]][next[1]][next[2]].HP <= 0 { // if we are at an end, let's save (LastMile 0 is an end because it does't need to be updated more than once, and therefore not a cycle)
				for j := len(interested) - 1; j >= 0; j-- { // append in reverse for hierarchy
					strand = append(strand, interested[j])
				}
				break
			} else {
				current = [3]int{next[0], next[1], next[2]}
			}
		}
	}

	return cycle, strand
}
