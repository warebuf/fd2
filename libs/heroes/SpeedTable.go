package heroes

import (
	"fmt"
)

func SpeedLookup(state [][][]*Hero, interest *Hero) int {

	if interest.Status != 1 || interest.LastMile != 2 {
		return interest.SPD
	}

	current := [3]int{interest.Index[0], interest.Index[1], interest.Index[2]}
	lowestSPD := interest.SPD

	visited := make(map[[3]int]struct{})
	visited[current] = struct{}{}

	for {

		if _, exists := visited[current]; exists {
			break
		}

		currentUnit := state[current[0]][current[1]][current[2]]

		if currentUnit.SPD < lowestSPD {
			lowestSPD = currentUnit.SPD
		}

		if currentUnit.Status != 1 || currentUnit.LastMile != 2 {
			break
		} else {
			next := [3]int{currentUnit.DynamicDestination[0], currentUnit.DynamicDestination[1], currentUnit.DynamicDestination[2]}
			if state[next[0]][next[1]][next[2]].HP > 0 {
				current = next
			} else {
				break
			}
		}

	}

	return lowestSPD

}

func SpeedTable(state [][][]*Hero, dependencies map[[3]int][][3]int) [][][]int {

	LM2Units := make(map[[3]int]bool) // find LM2 units, they are the units that need their speed checked
	table := [][][]int{}              // create speed table

	// create speed table and find the LM2 units
	for i := 0; i < len(state); i++ {
		subside := [][]int{}
		for j := 0; j < len(state[i]); j++ {
			subteam := []int{}
			for k := 0; k < len(state[i][j]); k++ {
				subteam = append(subteam, state[i][j][k].SPD)

				if state[i][j][k].Status == 1 && state[i][j][k].LastMile == 2 && state[i][j][k].HP > 0 {
					LM2Units[[3]int{i, j, k}] = false
				}
			}
			subside = append(subside, subteam)
		}
		table = append(table, subside)
	}

	fmt.Println("original SPD table", table)

	// determine the 0-degree LM2 units
	degree0 := [][3]int{}
	for LM2 := range LM2Units {
		isDegree0 := true
		if dependents, hasDependents := dependencies[LM2]; hasDependents {
			for i := 0; i < len(dependents); i++ {
				if state[dependents[i][0]][dependents[i][1]][dependents[i][2]].Status == 1 && state[dependents[i][0]][dependents[i][1]][dependents[i][2]].LastMile == 2 && state[dependents[i][0]][dependents[i][1]][dependents[i][2]].HP > 0 {
					isDegree0 = false
					break
				}
			}
		}
		if isDegree0 {
			degree0 = append(degree0, LM2)
			LM2Units[LM2] = true
		}

	}

	// set the speed for the LM2 strands
	for i := 0; i < len(degree0); i++ {

		// find the LM2 chain
		chain := [][3]int{degree0[i]}
		seen := make(map[[3]int]struct{})
		seen[degree0[i]] = struct{}{}
		cycleStart := [3]int{}
		isCycle := false
		current := [3]int(state[degree0[i][0]][degree0[i][1]][degree0[i][2]].DynamicDestination)
		for state[current[0]][current[1]][current[2]].Status == 1 && state[current[0]][current[1]][current[2]].LastMile == 2 && state[current[0]][current[1]][current[2]].HP > 0 {
			if _, found := seen[current]; found {
				cycleStart = current
				isCycle = true
				break
			}
			seen[current] = struct{}{}
			chain = append(chain, current)
			current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
		}

		// if the chain ends in a cycle, find the lowest SPD in the cycle, then solve upwards the chain
		cycleIndex := len(chain)
		if isCycle {
			for i := 0; i < len(chain); i++ {
				if chain[i] == cycleStart { // update the cycle with the lowest SPD
					cycleIndex = i

					lowestSPD := state[chain[i][0]][chain[i][1]][chain[i][2]].SPD // find the lowest SPD in the cycle
					for j := cycleIndex; j < len(chain); j++ {
						if state[chain[j][0]][chain[j][1]][chain[j][2]].SPD < lowestSPD {
							lowestSPD = state[chain[j][0]][chain[j][1]][chain[j][2]].SPD
						}
					}

					for j := cycleIndex; j < len(chain); j++ { // update the cycle
						table[chain[j][0]][chain[j][1]][chain[j][2]] = lowestSPD
						LM2Units[chain[j]] = true
					}
					break
				}
			}
		}

		// for the remaining strand, solve upwards by comparing the current vs the parent
		for i := cycleIndex - 1; i >= 0; i-- {
			toUpdate := chain[i]
			toUpdateSPD := table[toUpdate[0]][toUpdate[1]][toUpdate[2]]
			parent := [3]int(state[chain[i][0]][chain[i][1]][chain[i][2]].DynamicDestination)
			parentSPD := table[parent[0]][parent[1]][parent[2]]

			minSPD := toUpdateSPD
			if parentSPD < toUpdateSPD {
				minSPD = parentSPD
			}

			table[toUpdate[0]][toUpdate[1]][toUpdate[2]] = minSPD
			LM2Units[toUpdate] = true
		}

	}

	// for cycles, iterate through the cycle and find the lowest SPD unit (maybe set SPD to 0 idk)
	for LM2, completed := range LM2Units {
		if !completed && state[LM2[0]][LM2[1]][LM2[2]].HP > 0 {

			lowestSPD := state[LM2[0]][LM2[1]][LM2[2]].SPD
			current := [3]int(state[LM2[0]][LM2[1]][LM2[2]].DynamicDestination)
			for state[current[0]][current[1]][current[2]].Status == 1 && state[current[0]][current[1]][current[2]].LastMile == 2 && state[current[0]][current[1]][current[2]].HP > 0 {
				if current == LM2 {
					break
				}
				if state[current[0]][current[1]][current[2]].SPD < lowestSPD {
					lowestSPD = state[current[0]][current[1]][current[2]].SPD
				}
				current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
			}

			first := true
			current = LM2
			for {
				if !first && current == LM2 {
					break
				}
				table[current[0]][current[1]][current[2]] = lowestSPD
				LM2Units[current] = true
				first = false
				current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
			}
		}
	}

	fmt.Println("modified SPD table", table)

	return table
}
