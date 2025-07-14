package pathfinding

import (
	"container/heap"
	"fd/libs/maps"
	"fmt"
	"math"
)

type LPAKey struct {
	Location [2]float64
	H        float64
	RHS      float64
}

type LPAGraph struct {
	Start [2]float64
	Goal  [2]float64

	Nodes map[[2]float64][]float64 // [g, RHS]
	Edges map[[2]float64]map[[2]float64]float64

	PriorityQueue *LPAPQ
	Visited       map[[2]float64]bool
}

// LPAPQ represents the priority queue with an index map for efficient updates
type LPAPQ struct {
	queue    []LPAKey
	indexMap map[[2]float64]int // Maps node (coordinate) to its index in queue
}

func isLPAKeyLess(a, b LPAKey) bool {
	if a.H < b.H {
		return true
	} else if a.H > b.H {
		return false
	} else {
		if a.RHS < b.RHS {
			return true
		}
	}
	return false
}

func (pq *LPAPQ) Len() int { return len(pq.queue) }

func (pq *LPAPQ) Less(i, j int) bool {
	if pq.queue[i].H < pq.queue[j].H {
		return true
	} else if pq.queue[i].H > pq.queue[j].H {
		return false
	} else {
		return pq.queue[i].RHS < pq.queue[j].RHS
	}
}

func (pq *LPAPQ) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.indexMap[pq.queue[i].Location] = i
	pq.indexMap[pq.queue[j].Location] = j
}

func (pq *LPAPQ) Push(x interface{}) {
	node := x.(LPAKey)
	pq.queue = append(pq.queue, node)
	pq.indexMap[node.Location] = len(pq.queue) - 1
}

func (pq *LPAPQ) Pop() interface{} {
	old := pq.queue
	n := len(old)
	node := old[n-1]
	delete(pq.indexMap, node.Location) // Remove from index map
	pq.queue = old[:n-1]
	return node
}

func (pq *LPAPQ) topKey() LPAKey {
	if len(pq.queue) == 0 {
		return LPAKey{H: math.Inf(1), RHS: math.Inf(1)}
	}
	return pq.queue[0] // The top of the heap is always at index 0
}

func (pq *LPAPQ) Remove(location [2]float64) {
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

func (pq *LPAPQ) Update(location [2]float64, hCost float64, rhsCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, LPAKey{Location: location, H: hCost, RHS: rhsCost})
		return
	}

	pq.queue[index].H = hCost
	pq.queue[index].RHS = rhsCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func LPAInit(mp *maps.MapDetails, start [2]float64, goal [2]float64) *LPAGraph {

	masterLPA := &LPAGraph{
		Start: start,
		Goal:  goal,
		Nodes: make(map[[2]float64][]float64),
		Edges: make(map[[2]float64]map[[2]float64]float64),
		PriorityQueue: &LPAPQ{
			queue:    make([]LPAKey, 0),
			indexMap: make(map[[2]float64]int),
		},
		Visited: make(map[[2]float64]bool),
	}

	for i := 0; i < len(mp.Intersections); i++ {

		interestLocation := mp.Intersections[i]
		masterLPA.Nodes[interestLocation] = []float64{math.Inf(1), math.Inf(1)}

		indivMap := make(map[[2]float64]float64)
		for j := 0; j < len(mp.Edges[i]); j++ {
			edgeArray := mp.Intersections[mp.Edges[i][j]]
			edgeLocation := edgeArray
			indivMap[edgeLocation] = mp.EdgeLengths[i][j]
		}
		masterLPA.Edges[interestLocation] = indivMap
	}

	return masterLPA
}

func (graph *LPAGraph) CalculateKey(location [2]float64) LPAKey {
	H := maps.EuclideanDistance2(location, graph.Goal)
	G := graph.Nodes[location][0]
	RHS := graph.Nodes[location][1]
	return LPAKey{Location: location, H: math.Min(G, RHS) + H, RHS: math.Min(G, RHS)}
}

func (graph *LPAGraph) updateVertex(location [2]float64) {
	if location != graph.Start {
		minRHS := math.Inf(1)
		for neighborNode, edgeCost := range graph.Edges[location] {
			minRHS = math.Min(minRHS, graph.Nodes[neighborNode][0]+edgeCost)
		}
		graph.Nodes[location][1] = minRHS
	}
	if graph.Nodes[location][0] != graph.Nodes[location][1] {
		key := graph.CalculateKey(location)
		//fmt.Println("n", key)
		graph.PriorityQueue.Update(location, key.H, key.RHS)
	} else {
		graph.PriorityQueue.Remove(location)
	}
}

func (graph *LPAGraph) AddNode(location [2]float64) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = []float64{math.Inf(1), math.Inf(1)}
		graph.Edges[location] = make(map[[2]float64]float64)
	}

}

func (graph *LPAGraph) AddEdge(from [2]float64, to [2]float64, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = []float64{math.Inf(1), math.Inf(1)}
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

func (graph *LPAGraph) UpdateEdgeCost(from [2]float64, to [2]float64, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Edges[from]; !exists {
		graph.Edges[from] = make(map[[2]float64]float64)
	}

	if _, exists := graph.Edges[to]; !exists {
		graph.Edges[to] = make(map[[2]float64]float64)
	}

	graph.Edges[from][to] = cost
	graph.Edges[to][from] = cost

	graph.updateVertex(from)
	graph.updateVertex(to)

	graph.ComputeShortestPath()
}

func (graph *LPAGraph) ComputeShortestPath() {

	for isLPAKeyLess(graph.PriorityQueue.topKey(), graph.CalculateKey(graph.Goal)) || graph.Nodes[graph.Goal][0] != graph.Nodes[graph.Goal][1] {

		if len(graph.PriorityQueue.queue) == 0 {
			break
		}

		current := heap.Pop(graph.PriorityQueue).(LPAKey)
		//fmt.Println("current", current.Location, current.H, current.RHS, len(graph.PriorityQueue.queue), len(graph.Visited))

		if graph.Nodes[current.Location][0] > graph.Nodes[current.Location][1] {
			graph.Nodes[current.Location][0] = graph.Nodes[current.Location][1]
			graph.Visited[current.Location] = true
		} else {
			graph.Nodes[current.Location][0] = math.Inf(1)
			graph.updateVertex(current.Location)
		}

		for neighborLocation, _ := range graph.Edges[current.Location] {
			_, exists := graph.Visited[neighborLocation]
			if exists == false {
				// TODO: I feel that this call is inefficient
				// in this instance, we should iterate through the neighbors, and update their RHS value/H value IF
				// our newly found confirmed g-value + edge cost is better
				// instead of iterating through all the neighbors
				// I believe this property holds for all instances where we call update Vertex
				// if it doesn't we can just make two versions of updateVertex
				graph.updateVertex(neighborLocation)
			}
		}
	}
}

func (graph *LPAGraph) ReturnShortestPath() [][2]float64 {

	var ans [][2]float64

	if graph.Nodes[graph.Goal][0] == math.Inf(1) {
		fmt.Println("the path does not exist")
		return ans
	}

	current := graph.Goal
	for current != graph.Start {
		ans = append(ans, current)

		var smallestNeighbor [2]float64
		found := false
		smallest := math.Inf(1)

		for neighborLocation, neighborEdge := range graph.Edges[current] {
			if graph.Nodes[neighborLocation][0]+neighborEdge < smallest {
				found = true
				smallest = graph.Nodes[neighborLocation][0] + neighborEdge
				smallestNeighbor = neighborLocation
			}
		}

		if !found {
			fmt.Println("the path terminates early")
			return [][2]float64{}
		}
		current = smallestNeighbor
	}
	ans = append(ans, current)
	var ans2 [][2]float64
	for i := len(ans) - 1; i >= 0; i-- {
		ans2 = append(ans2, ans[i])
	}
	fmt.Println(ans2)
	sum := 0.0
	for i := 0; i < len(ans)-1; i++ {
		sum += maps.EuclideanDistance2(ans[i], ans[i+1])
	}
	fmt.Println("sum", sum)

	return ans2

}
