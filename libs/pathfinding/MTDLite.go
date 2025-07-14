package pathfinding

import (
	"container/heap"
	"fd/libs/maps"
	"fmt"
	"math"
)

type MTDLiteKey struct {
	Location [2]float64
	H        float64
	RHS      float64
}

type MTDLiteGraph struct {
	Start [2]float64
	Goal  [2]float64
	Last  [2]float64

	Nodes map[[2]float64][]float64 // [g, RHS]
	Edges map[[2]float64]map[[2]float64]float64

	PriorityQueue *MTDLitePQ
	Visited       map[[2]float64]bool

	Km float64
}

type MTDLitePQ struct {
	queue    []MTDLiteKey
	indexMap map[[2]float64]int // Maps node (coordinate) to its index in queue
}

func isMTDLiteKeyLess(a, b MTDLiteKey) bool {
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

func (pq *MTDLitePQ) Len() int { return len(pq.queue) }

func (pq *MTDLitePQ) Less(i, j int) bool {
	if pq.queue[i].H < pq.queue[j].H {
		return true
	} else if pq.queue[i].H > pq.queue[j].H {
		return false
	} else {
		return pq.queue[i].RHS < pq.queue[j].RHS
	}
}

func (pq *MTDLitePQ) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.indexMap[pq.queue[i].Location] = i
	pq.indexMap[pq.queue[j].Location] = j
}

func (pq *MTDLitePQ) Push(x interface{}) {
	node := x.(MTDLiteKey)
	pq.queue = append(pq.queue, node)
	pq.indexMap[node.Location] = len(pq.queue) - 1
}

func (pq *MTDLitePQ) Pop() interface{} {
	old := pq.queue
	n := len(old)
	node := old[n-1]
	delete(pq.indexMap, node.Location) // Remove from index map
	pq.queue = old[:n-1]
	return node
}

func (pq *MTDLitePQ) topKey() MTDLiteKey {
	if len(pq.queue) == 0 {
		return MTDLiteKey{H: math.Inf(1), RHS: math.Inf(1)}
	}
	return pq.queue[0] // The top of the heap is always at index 0
}

func (pq *MTDLitePQ) Remove(location [2]float64) {
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

func (pq *MTDLitePQ) Update(location [2]float64, hCost float64, rhsCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, MTDLiteKey{Location: location, H: hCost, RHS: rhsCost})
		return
	}

	pq.queue[index].H = hCost
	pq.queue[index].RHS = rhsCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func MTDLiteInit(mp *maps.MapDetails, start [2]float64, goal [2]float64) *MTDLiteGraph {

	masterDLite := &MTDLiteGraph{
		Start: start,
		Goal:  goal,
		Last:  start,
		Nodes: make(map[[2]float64][]float64),
		Edges: make(map[[2]float64]map[[2]float64]float64),
		PriorityQueue: &MTDLitePQ{
			queue:    make([]MTDLiteKey, 0),
			indexMap: make(map[[2]float64]int),
		},
		Visited: make(map[[2]float64]bool),
		Km:      0.0,
	}

	for i := 0; i < len(mp.Intersections); i++ {

		interestLocation := mp.Intersections[i]
		masterDLite.Nodes[interestLocation] = []float64{math.Inf(1), math.Inf(1)}

		indivMap := make(map[[2]float64]float64)
		for j := 0; j < len(mp.Edges[i]); j++ {
			edgeArray := mp.Intersections[mp.Edges[i][j]]
			edgeLocation := edgeArray
			indivMap[edgeLocation] = mp.EdgeLengths[i][j]
		}
		masterDLite.Edges[interestLocation] = indivMap
	}

	return masterDLite
}

func (graph *MTDLiteGraph) CalculateKey(location [2]float64) MTDLiteKey {
	H := maps.EuclideanDistance2(location, graph.Start)
	G := graph.Nodes[location][0]
	RHS := graph.Nodes[location][1]
	return MTDLiteKey{Location: location, H: math.Min(G, RHS) + H + graph.Km, RHS: math.Min(G, RHS)}
}

func (graph *MTDLiteGraph) updateVertex(location [2]float64) {
	if location != graph.Goal {
		minRHS := math.Inf(1)
		for neighborNode, edgeCost := range graph.Edges[location] {
			minRHS = math.Min(minRHS, graph.Nodes[neighborNode][0]+edgeCost)
		}
		graph.Nodes[location][1] = minRHS
	}
	if graph.Nodes[location][0] != graph.Nodes[location][1] {
		key := graph.CalculateKey(location)
		//fmt.Println("edge", key)
		graph.PriorityQueue.Update(location, key.H, key.RHS)
	} else {
		graph.PriorityQueue.Remove(location)
	}
}

func (graph *MTDLiteGraph) AddNode(location [2]float64) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = []float64{math.Inf(1), math.Inf(1)}
		graph.Edges[location] = make(map[[2]float64]float64)
	}

}

func (graph *MTDLiteGraph) AddEdge(from [2]float64, to [2]float64, cost float64) {

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

func (graph *MTDLiteGraph) UpdateEdgeCost(from [2]float64, to [2]float64, cost float64) {

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

func (graph *MTDLiteGraph) ComputeShortestPath() {

	for isMTDLiteKeyLess(graph.PriorityQueue.topKey(), graph.CalculateKey(graph.Start)) || graph.Nodes[graph.Start][0] != graph.Nodes[graph.Start][1] {

		if len(graph.PriorityQueue.queue) == 0 {
			break
		}

		oldKey := heap.Pop(graph.PriorityQueue).(MTDLiteKey)
		newKey := graph.CalculateKey(oldKey.Location)
		//fmt.Println("current", newKey, len(graph.PriorityQueue.queue), len(graph.Visited))

		if isMTDLiteKeyLess(oldKey, newKey) {
			graph.PriorityQueue.Update(newKey.Location, newKey.H, newKey.RHS)
		} else if graph.Nodes[oldKey.Location][0] > graph.Nodes[oldKey.Location][1] {
			graph.Nodes[oldKey.Location][0] = graph.Nodes[oldKey.Location][1]
			graph.Visited[oldKey.Location] = true
			for neighborLocation, _ := range graph.Edges[oldKey.Location] {
				_, exists := graph.Visited[neighborLocation]
				if !exists {
					graph.updateVertex(neighborLocation)
				}
			}
		} else {
			graph.Nodes[oldKey.Location][0] = math.Inf(1)
			graph.updateVertex(oldKey.Location)
			for neighborLocation := range graph.Edges[oldKey.Location] {
				_, exists := graph.Visited[neighborLocation]
				if !exists {
					graph.updateVertex(neighborLocation)
				}
			}
		}
	}

}

func (graph *MTDLiteGraph) ReturnShortestPath() [][2]float64 {

	var ans [][2]float64
	visited := make(map[[2]float64]bool)

	if graph.Nodes[graph.Goal][0] == math.Inf(1) {
		fmt.Println("the path does not exist")
		return ans
	}

	current := graph.Start

	for current != graph.Goal {
		ans = append(ans, current)
		visited[current] = true

		var smallestNeighbor *[2]float64
		smallest := math.Inf(1)

		for neighborLocation, neighborEdge := range graph.Edges[current] {
			_, found := visited[neighborLocation]
			if !found && graph.Nodes[neighborLocation][0]+neighborEdge < smallest {
				smallest = graph.Nodes[neighborLocation][0] + neighborEdge
				smallestNeighbor = &neighborLocation
			}
		}

		if smallestNeighbor == nil {
			fmt.Println("the path terminates early")
			return [][2]float64{}
		}
		current = *smallestNeighbor
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
