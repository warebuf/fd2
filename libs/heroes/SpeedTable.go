package heroes

import "fmt"

func SpeedLookup(state [][][]*Hero, interest *Hero) int {

	if interest.Status != 1 || interest.LastMile != 2 {
		return interest.SPD
	}

	current := interest
	lowestSPD := current.SPD
	first := true

	for {

		if !first && current == interest {
			break
		}

		if current.SPD < lowestSPD {
			lowestSPD = current.SPD
		}

		if current.Status != 1 || current.LastMile != 2 {
			break
		} else {
			next := state[current.DynamicDestination[0]][current.DynamicDestination[1]][current.DynamicDestination[2]]
			if next.HP > 0 {
				current = next
			} else {
				break
			}
		}
		first = false

	}

	return lowestSPD

}

func SpeedTable(state [][][]*Hero, dependencies map[[3]int][][3]int) [][][]int {

	LM2Units := make(map[[3]int]bool) // find LM2 units
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

		// find the lowest SPD in this LM2 strand
		lowestSPD := state[degree0[i][0]][degree0[i][1]][degree0[i][2]].SPD
		current := [3]int(state[degree0[i][0]][degree0[i][1]][degree0[i][2]].DynamicDestination)
		for state[current[0]][current[1]][current[2]].Status == 1 && state[current[0]][current[1]][current[2]].LastMile == 2 && state[current[0]][current[1]][current[2]].HP > 0 {
			if state[current[0]][current[1]][current[2]].SPD < lowestSPD {
				lowestSPD = state[current[0]][current[1]][current[2]].SPD
			}
			current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
		}

		// check if the head (not a LM2 unit) is lowest SPD
		if state[current[0]][current[1]][current[2]].HP > 0 && state[current[0]][current[1]][current[2]].SPD < lowestSPD {
			lowestSPD = state[current[0]][current[1]][current[2]].SPD
		}

		// update the LM2 strand with the lowest SPD
		current = degree0[i]
		for state[current[0]][current[1]][current[2]].Status == 1 && state[current[0]][current[1]][current[2]].LastMile == 2 && state[current[0]][current[1]][current[2]].HP > 0 {
			table[current[0]][current[1]][current[2]] = lowestSPD
			current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
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
				first = false
				current = [3]int(state[current[0]][current[1]][current[2]].DynamicDestination)
			}
		}
	}

	fmt.Println("modified SPD table", table)

	return table
}
