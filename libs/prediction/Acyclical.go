package prediction

import (
	"fd/libs/heroes"
)

// this problem is a well known problem called path decomposition in functional graphs
// each node has 0 or 1 outgoing directed edges, we want to find all strands/cycles

func GetStrandsAndCycles(state [][][]*heroes.Hero) ([][][3]int, [][][3]int) {
	visited := make(map[[3]int]int) // 0 = unvisited, 1 = currently visiting, 2 = visited
	cycles := [][][3]int{}
	strands := [][][3]int{}

	// first find all starting nodes
	dependents := make(map[[3]int]int)
	for i := 0; i < len(state); i++ {
		for j := 0; j < len(state[i]); j++ {
			for k := 0; k < len(state[i][j]); k++ {
				visited[[3]int{i, j, k}] = 0
				if len(state[i][j][k].DynamicDestination) > 0 {
					to := [3]int{state[i][j][k].DynamicDestination[0], state[i][j][k].DynamicDestination[1], state[i][j][k].DynamicDestination[2]}
					dependents[to] += 1
				}
			}
		}
	}

	// traverse the starting nodes, record all cycles
	for key, val := range dependents {

		if val != 0 {
			continue
		}

		path := [][3]int{}
		current := key
		for {
			if visited[current] == 1 { // cycle
				cycleStart := -1
				for idx, n := range path {
					if n == current {
						cycleStart = idx
						break
					}
				}
				cycles = append(cycles, path[cycleStart:])
				strands = append(strands, path)
				break
			} else if visited[current] == 2 {
				strands = append(strands, path) // already visited, no need to continue
				break
			}

			visited[current] = 1
			path = append(path, current)

			dest := state[current[0]][current[1]][current[2]].DynamicDestination
			if len(dest) < 3 { // deadend
				strands = append(strands, path)
				break
			}
			current = [3]int{dest[0], dest[1], dest[2]}
		}

		// Mark all as fully visited
		for _, p := range path {
			visited[p] = 2
		}

	}

	// we have some cycles left, do the same to all
	for key, val := range visited {

		if val != 0 {
			continue
		}

		path := [][3]int{}
		current := key
		for {
			if visited[current] == 1 { // cycle
				cycleStart := -1
				for idx, n := range path {
					if n == current {
						cycleStart = idx
						break
					}
				}
				cycles = append(cycles, path[cycleStart:])
				strands = append(strands, path)
				break
			} else if visited[current] == 2 {
				strands = append(strands, path) // already visited, no need to continue
				break
			}

			visited[current] = 1
			path = append(path, current)

			dest := state[current[0]][current[1]][current[2]].DynamicDestination
			if len(dest) < 3 { // deadend
				strands = append(strands, path)
				break
			}
			current = [3]int{dest[0], dest[1], dest[2]}
		}

		// Mark all as fully visited
		for _, p := range path {
			visited[p] = 2
		}

	}

	return cycles, strands
}

// now given cycles and strands
// 1. calculate the best path for cycles
// 2. go down the strand and calculate up
