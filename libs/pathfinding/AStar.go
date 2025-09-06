package pathfinding

import (
	"container/heap"
	"fd/libs/bezier"
	"fd/libs/maps"
	"math"
)

type AStarKey struct {
	Location [2]float64
	F        float64
}

type AStarGraph struct {
	Start [2]float64
	Goal  [2]float64

	Nodes  map[[2]float64]float64 // g-score
	Edges  map[[2]float64]map[[2]float64]float64
	Parent map[[2]float64]*[2]float64

	PriorityQueue *AStarPQ
	Visited       map[[2]float64]bool
}

type AStarPQ struct {
	queue    []AStarKey
	indexMap map[[2]float64]int // Maps node (coordinate) to its index in queue
}

func (pq *AStarPQ) Len() int { return len(pq.queue) }

func (pq *AStarPQ) Less(i, j int) bool {
	return pq.queue[i].F < pq.queue[j].F
}

func (pq *AStarPQ) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.indexMap[pq.queue[i].Location] = i
	pq.indexMap[pq.queue[j].Location] = j
}

func (pq *AStarPQ) Push(x interface{}) {
	node := x.(AStarKey)
	pq.queue = append(pq.queue, node)
	pq.indexMap[node.Location] = len(pq.queue) - 1
}

func (pq *AStarPQ) Pop() interface{} {
	old := pq.queue
	n := len(old)
	node := old[n-1]
	delete(pq.indexMap, node.Location) // Remove from index map
	pq.queue = old[:n-1]
	return node
}

func (pq *AStarPQ) Remove(location [2]float64) {
	removeIndex, exists := pq.indexMap[location]
	if !exists {
		return
	}
	lastIndex := len(pq.queue) - 1
	pq.Swap(removeIndex, lastIndex)
	pq.queue = pq.queue[:lastIndex]
	delete(pq.indexMap, location)
	if removeIndex < len(pq.queue) {
		heap.Fix(pq, removeIndex)
	}
}

func (pq *AStarPQ) Update(location [2]float64, fCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, AStarKey{Location: location, F: fCost})
		return
	}

	pq.queue[index].F = fCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func AStarInit(mp *maps.MapDetails, start [2]float64, goal [2]float64) *AStarGraph {

	masterAStar := &AStarGraph{
		Start: start,
		Goal:  goal,

		Nodes:  make(map[[2]float64]float64),
		Edges:  make(map[[2]float64]map[[2]float64]float64),
		Parent: make(map[[2]float64]*[2]float64),

		PriorityQueue: &AStarPQ{
			queue:    make([]AStarKey, 0),
			indexMap: make(map[[2]float64]int),
		},
		Visited: make(map[[2]float64]bool),
	}

	for i := 0; i < len(mp.Intersections); i++ {

		interestLocation := mp.Intersections[i]
		masterAStar.Nodes[interestLocation] = math.Inf(1)

		indivMap := make(map[[2]float64]float64)
		for j := 0; j < len(mp.Edges[i]); j++ {
			edgeArray := mp.Intersections[mp.Edges[i][j]]
			indivMap[edgeArray] = mp.EdgeLengths[i][j]
		}
		masterAStar.Edges[interestLocation] = indivMap

		masterAStar.Parent[interestLocation] = nil
	}

	if _, exists := masterAStar.Nodes[start]; !exists {
		edgeIndex, edgeCost := masterAStar.getEdges(start, mp)
		masterAStar.AddNode(start)
		for i := 0; i < len(edgeIndex); i++ {
			masterAStar.AddEdge(start, edgeIndex[i], edgeCost[i])
		}
	}

	if _, exists := masterAStar.Nodes[goal]; !exists {
		edgeIndex, edgeCost := masterAStar.getEdges(goal, mp)
		masterAStar.AddNode(goal)
		for i := 0; i < len(edgeIndex); i++ {
			masterAStar.AddEdge(goal, edgeIndex[i], edgeCost[i])
		}
	}

	masterAStar.Nodes[start] = 0
	heap.Push(masterAStar.PriorityQueue, masterAStar.CalculateKey(start))

	return masterAStar
}

func (graph *AStarGraph) CalculateKey(location [2]float64) AStarKey {
	F := graph.Nodes[location] + maps.EuclideanDistance2(location, graph.Goal)
	return AStarKey{Location: location, F: F}
}

func (graph *AStarGraph) AddNode(location [2]float64) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = math.Inf(1)
		graph.Edges[location] = make(map[[2]float64]float64)
	}

}

func (graph *AStarGraph) AddEdge(from [2]float64, to [2]float64, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = math.Inf(1)
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = math.Inf(1)
	}

	if _, exists := graph.Edges[from]; !exists {
		graph.Edges[from] = make(map[[2]float64]float64)
	}

	if _, exists := graph.Edges[to]; !exists {
		graph.Edges[to] = make(map[[2]float64]float64)
	}

	graph.Edges[from][to] = cost
	graph.Edges[to][from] = cost

}

func (graph *AStarGraph) ComputeShortestPath() [][2]float64 {

	count := 0

	for graph.PriorityQueue.Len() > 0 {

		current := heap.Pop(graph.PriorityQueue).(AStarKey)

		if current.Location == graph.Goal {

			ans := [][2]float64{}

			currentPath := &graph.Goal

			for currentPath != nil {

				ans = append(ans, *currentPath)
				currentPath = graph.Parent[*currentPath]

			}

			ans2 := [][2]float64{}
			for i := len(ans) - 1; i >= 0; i-- {
				ans2 = append(ans2, ans[i])
			}

			sum := 0.0
			for i := 0; i < len(ans)-1; i++ {
				sum += maps.EuclideanDistance2(ans[i], ans[i+1])
			}

			return ans2
		}

		graph.Visited[current.Location] = true
		count += 1

		for neighbor, cost := range graph.Edges[current.Location] {

			_, exists := graph.Visited[neighbor]
			if exists {
				continue
			} else {
				tentativeG := graph.Nodes[current.Location] + cost

				if tentativeG < graph.Nodes[neighbor] {
					graph.Parent[neighbor] = &current.Location
					graph.Nodes[neighbor] = tentativeG
					newF := graph.CalculateKey(neighbor)
					graph.PriorityQueue.Update(neighbor, newF.F) // I thought you would update only if G val is better

				}

			}

		}

	}

	return [][2]float64{}

}

func (graph *AStarGraph) getEdges(c1 [2]float64, mp *maps.MapDetails) ([][2]float64, []float64) {
	edges := [][2]float64{}
	dist := make([]float64, 0)

	for key := range graph.Nodes {
		line_seg := []float64{key[0], key[1], c1[0], c1[1]}
		blocked := false
		for j := 0; j < len(mp.Obstacles); j++ {
			if mp.Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mp.Obstacles[j], line_seg) {
					blocked = true
					break
				}
			}
		}
		if !blocked {
			edges = append(edges, key)
			dist = append(dist, maps.EuclideanDistance2(c1, key))
		}
	}

	return edges, dist
}
