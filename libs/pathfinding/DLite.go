package pathfinding

import (
	"container/heap"
	"fd/libs/maps"
	"fmt"
	"math"
)

type DLiteKey struct {
	Location Location
	H        float64
	RHS      float64
}

type DLiteGraph struct {
	Start Location
	Goal  Location
	Last  Location

	Nodes map[Location][]float64 // [g, RHS]
	Edges map[Location]map[Location]float64

	PriorityQueue *DLitePQ
	Visited       map[Location]bool

	Km float64
}

type DLitePQ struct {
	queue    []DLiteKey
	indexMap map[Location]int // Maps node (coordinate) to its index in queue
}

func isDLiteKeyLess(a, b DLiteKey) bool {
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

func (pq *DLitePQ) Len() int { return len(pq.queue) }

func (pq *DLitePQ) Less(i, j int) bool {
	if pq.queue[i].H < pq.queue[j].H {
		return true
	} else if pq.queue[i].H > pq.queue[j].H {
		return false
	} else {
		return pq.queue[i].RHS < pq.queue[j].RHS
	}
}

func (pq *DLitePQ) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.indexMap[pq.queue[i].Location] = i
	pq.indexMap[pq.queue[j].Location] = j
}

func (pq *DLitePQ) Push(x interface{}) {
	node := x.(DLiteKey)
	pq.queue = append(pq.queue, node)
	pq.indexMap[node.Location] = len(pq.queue) - 1
}

func (pq *DLitePQ) Pop() interface{} {
	old := pq.queue
	n := len(old)
	node := old[n-1]
	delete(pq.indexMap, node.Location) // Remove from index map
	pq.queue = old[:n-1]
	return node
}

func (pq *DLitePQ) topKey() DLiteKey {
	if len(pq.queue) == 0 {
		return DLiteKey{H: math.Inf(1), RHS: math.Inf(1)}
	}
	return pq.queue[0] // The top of the heap is always at index 0
}

func (pq *DLitePQ) Remove(location Location) {
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

func (pq *DLitePQ) Update(location Location, hCost float64, rhsCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, DLiteKey{Location: location, H: hCost, RHS: rhsCost})
		return
	}

	pq.queue[index].H = hCost
	pq.queue[index].RHS = rhsCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func DLiteInit(mp *maps.MapDetails, start Location, goal Location) *DLiteGraph {

	masterDLite := &DLiteGraph{
		Start: start,
		Goal:  goal,
		Last:  start,
		Nodes: make(map[Location][]float64),
		Edges: make(map[Location]map[Location]float64),
		PriorityQueue: &DLitePQ{
			queue:    make([]DLiteKey, 0),
			indexMap: make(map[Location]int),
		},
		Visited: make(map[Location]bool),
		Km:      0.0,
	}

	for i := 0; i < len(mp.Intersections); i++ {

		interestLocation := Location{X: mp.Intersections[i][0], Y: mp.Intersections[i][1]}
		masterDLite.Nodes[interestLocation] = []float64{math.Inf(1), math.Inf(1)}

		indivMap := make(map[Location]float64)
		for j := 0; j < len(mp.Intersections_edges_int[i]); j++ {
			edgeArray := mp.Intersections[mp.Intersections_edges_int[i][j]]
			edgeLocation := Location{edgeArray[0], edgeArray[1]}
			indivMap[edgeLocation] = mp.Intersections_edgelength[i][j]
		}
		masterDLite.Edges[interestLocation] = indivMap
	}

	return masterDLite
}

func (graph *DLiteGraph) CalculateKey(location Location) DLiteKey {
	H := maps.EuclideanDistance(location.X, location.Y, graph.Start.X, graph.Start.Y)
	G := graph.Nodes[location][0]
	RHS := graph.Nodes[location][1]
	return DLiteKey{Location: location, H: math.Min(G, RHS) + H + graph.Km, RHS: math.Min(G, RHS)}
}

func (graph *DLiteGraph) updateVertex(location Location) {
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

func (graph *DLiteGraph) AddNode(location Location) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = []float64{math.Inf(1), math.Inf(1)}
		graph.Edges[location] = make(map[Location]float64)
	}

}

func (graph *DLiteGraph) AddEdge(from Location, to Location, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Edges[from]; !exists {
		graph.Edges[from] = make(map[Location]float64)
	}

	if _, exists := graph.Edges[to]; !exists {
		graph.Edges[to] = make(map[Location]float64)
	}

	graph.Edges[from][to] = cost
	graph.Edges[to][from] = cost

}

func (graph *DLiteGraph) UpdateEdgeCost(from Location, to Location, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = []float64{math.Inf(1), math.Inf(1)}
	}

	if _, exists := graph.Edges[from]; !exists {
		graph.Edges[from] = make(map[Location]float64)
	}

	if _, exists := graph.Edges[to]; !exists {
		graph.Edges[to] = make(map[Location]float64)
	}

	graph.Edges[from][to] = cost
	graph.Edges[to][from] = cost

	graph.updateVertex(from)
	graph.updateVertex(to)

	graph.ComputeShortestPath()
}

func (graph *DLiteGraph) ComputeShortestPath() {

	for isDLiteKeyLess(graph.PriorityQueue.topKey(), graph.CalculateKey(graph.Start)) || graph.Nodes[graph.Start][0] != graph.Nodes[graph.Start][1] {

		if len(graph.PriorityQueue.queue) == 0 {
			break
		}

		oldKey := heap.Pop(graph.PriorityQueue).(DLiteKey)
		newKey := graph.CalculateKey(oldKey.Location)
		//fmt.Println("current", newKey, len(graph.PriorityQueue.queue), len(graph.Visited))

		if isDLiteKeyLess(oldKey, newKey) {
			graph.PriorityQueue.Update(newKey.Location, newKey.H, newKey.RHS)
		} else if graph.Nodes[oldKey.Location][0] > graph.Nodes[oldKey.Location][1] {
			graph.Nodes[oldKey.Location][0] = graph.Nodes[oldKey.Location][1]
			graph.Visited[oldKey.Location] = true
			for neighborLocation, _ := range graph.Edges[oldKey.Location] {
				_, exists := graph.Visited[neighborLocation]
				if exists == false {
					graph.updateVertex(neighborLocation)
				}
			}
		} else {
			graph.Nodes[oldKey.Location][0] = math.Inf(1)
			graph.updateVertex(oldKey.Location)
			for neighborLocation, _ := range graph.Edges[oldKey.Location] {
				_, exists := graph.Visited[neighborLocation]
				if exists == false {
					graph.updateVertex(neighborLocation)
				}
			}
		}
	}

}

func (graph *DLiteGraph) ReturnShortestPath() []Location {

	var ans []Location
	visited := make(map[Location]bool)

	if graph.Nodes[graph.Goal][0] == math.Inf(1) {
		fmt.Println("the path does not exist")
		return ans
	}

	current := graph.Start

	for current != graph.Goal {
		ans = append(ans, current)
		visited[current] = true

		var smallestNeighbor *Location
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
			return []Location{}
		}
		current = *smallestNeighbor
	}
	ans = append(ans, current)
	var ans2 []Location
	for i := len(ans) - 1; i >= 0; i-- {
		ans2 = append(ans2, ans[i])
	}
	fmt.Println(ans2)
	sum := 0.0
	for i := 0; i < len(ans)-1; i++ {
		sum += maps.EuclideanDistance(ans[i].X, ans[i].Y, ans[i+1].X, ans[i+1].Y)
	}
	fmt.Println("sum", sum)

	return ans2

}
