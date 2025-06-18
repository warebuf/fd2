package prediction

import (
	"fd/libs/heroes"
)

// something like this
func Acyclical(state [][][]*heroes.Hero) {

	inDegree := make(map[int]int)
	visited := make(map[int]bool)
	nodes := make(map[int]bool)

	// Compute in-degrees and collect all nodes
	for from, to := range edges {
		inDegree[to]++
		nodes[from] = true
		nodes[to] = true
	}

	strands := [][]int{}

	// First: explore all paths starting from nodes with in-degree 0
	for node := range nodes {
		if inDegree[node] != 0 || visited[node] {
			continue
		}
		path := []int{}
		curr := node
		for !visited[curr] {
			path = append(path, curr)
			visited[curr] = true
			next, exists := edges[curr]
			if !exists {
				break
			}
			curr = next
		}
		strands = append(strands, path)
	}

	// Second: visit any remaining unvisited nodes (i.e. part of cycles)
	for node := range nodes {
		if visited[node] {
			continue
		}
		path := []int{}
		curr := node
		for !visited[curr] {
			path = append(path, curr)
			visited[curr] = true
			next, exists := edges[curr]
			if !exists {
				break
			}
			curr = next
		}
		strands = append(strands, path)
	}

	return strands

}
