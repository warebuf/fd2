package pathfinding

import (
	"container/heap"
	"fd/libs/bezier"
	"fd/libs/maps"
	"fmt"
	"math"
	"time"
)

type DLiteKey struct {
	Location [2]float64
	H        float64
	RHS      float64
}

type DLiteGraph struct {
	Start [2]float64
	Goal  [2]float64
	Last  [2]float64

	Nodes map[[2]float64][]float64 // [g, RHS]
	Edges map[[2]float64]map[[2]float64]float64

	PriorityQueue *DLitePQ
	Visited       map[[2]float64]bool

	Km float64

	TemporaryLocations map[[2]float64]bool
}

type DLitePQ struct {
	queue    []DLiteKey
	indexMap map[[2]float64]int // Maps node (coordinate) to its index in queue
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

func (pq *DLitePQ) Remove(location [2]float64) {
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

func (pq *DLitePQ) Update(location [2]float64, hCost float64, rhsCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, DLiteKey{Location: location, H: hCost, RHS: rhsCost})
		return
	}

	pq.queue[index].H = hCost
	pq.queue[index].RHS = rhsCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func DLiteInit(mp *maps.MapDetails, start [2]float64, goal [2]float64) *DLiteGraph {

	masterDLite := &DLiteGraph{
		Start: start,
		Goal:  goal,
		Last:  start,
		Nodes: make(map[[2]float64][]float64),
		Edges: make(map[[2]float64]map[[2]float64]float64),
		PriorityQueue: &DLitePQ{
			queue:    make([]DLiteKey, 0),
			indexMap: make(map[[2]float64]int),
		},
		Visited:            make(map[[2]float64]bool),
		Km:                 0.0,
		TemporaryLocations: make(map[[2]float64]bool),
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

func (graph *DLiteGraph) CalculateKey(location [2]float64) DLiteKey {
	H := maps.EuclideanDistance2(location, graph.Start)
	G := graph.Nodes[location][0]
	RHS := graph.Nodes[location][1]
	return DLiteKey{Location: location, H: math.Min(G, RHS) + H + graph.Km, RHS: math.Min(G, RHS)}
}

func (graph *DLiteGraph) updateVertex(location [2]float64) {
	if location != graph.Goal {
		minRHS := math.Inf(1)
		for neighborNode, edgeCost := range graph.Edges[location] {
			//fmt.Println(neighborNode, graph.Nodes[neighborNode][0], edgeCost)
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

func (graph *DLiteGraph) AddNode(location [2]float64) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = []float64{math.Inf(1), math.Inf(1)}
		graph.Edges[location] = make(map[[2]float64]float64)

		graph.TemporaryLocations[location] = true
	}

}

func (graph *DLiteGraph) AddEdge(from [2]float64, to [2]float64, cost float64) {

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

func (graph *DLiteGraph) UpdateEdgeCost(from [2]float64, to [2]float64, cost float64) {

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

func (graph *DLiteGraph) ComputeShortestPath() {

	count := 0
	count2 := 0

	for isDLiteKeyLess(graph.PriorityQueue.topKey(), graph.CalculateKey(graph.Start)) || graph.Nodes[graph.Start][0] != graph.Nodes[graph.Start][1] {

		if len(graph.PriorityQueue.queue) == 0 {
			break
		}

		oldKey := heap.Pop(graph.PriorityQueue).(DLiteKey)
		newKey := graph.CalculateKey(oldKey.Location)
		//fmt.Println("current", newKey, graph.Nodes[newKey.Location], len(graph.PriorityQueue.queue), len(graph.Visited))

		count2 += 1

		if isDLiteKeyLess(oldKey, newKey) {
			graph.PriorityQueue.Update(newKey.Location, newKey.H, newKey.RHS)
		} else if graph.Nodes[oldKey.Location][0] > graph.Nodes[oldKey.Location][1] {
			count += 1

			graph.Nodes[oldKey.Location][0] = graph.Nodes[oldKey.Location][1]
			graph.Visited[oldKey.Location] = true
			for neighborLocation, _ := range graph.Edges[oldKey.Location] {
				_, exists := graph.Visited[neighborLocation]
				//fmt.Println("neighbor", neighborLocation, graph.Nodes[neighborLocation])
				if exists == false {
					graph.updateVertex(neighborLocation)
				}
			}
		} else {
			count += 1

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

	fmt.Println("D* count", count2, count)

}

func (graph *DLiteGraph) ReturnShortestPath() [][2]float64 {

	var path [][2]float64
	visited := make(map[[2]float64]bool)

	if graph.Nodes[graph.Goal][0] == math.Inf(1) {
		fmt.Println("the path does not exist")
		return path
	}

	current := graph.Last

	for current != graph.Goal {
		path = append(path, current)
		visited[current] = true

		var smallestNeighbor [2]float64
		atLeastOneFound := false
		smallest := math.Inf(1)

		for neighborLocation, neighborEdge := range graph.Edges[current] {
			_, found := visited[neighborLocation]
			if !found && graph.Nodes[neighborLocation][0]+neighborEdge < smallest {
				atLeastOneFound = true
				smallest = graph.Nodes[neighborLocation][0] + neighborEdge
				smallestNeighbor = neighborLocation
			}
		}

		if !atLeastOneFound {
			fmt.Println("the path terminates early")
			return [][2]float64{}
		}
		current = smallestNeighbor
	}
	path = append(path, current)
	sum := 0.0
	for i := 0; i < len(path)-1; i++ {
		sum += maps.EuclideanDistance2(path[i], path[i+1])
	}
	fmt.Println("D* LITE path", path)
	fmt.Println("D* LITE sum", sum)

	return path

}

func (graph *DLiteGraph) GetEdges(mp *maps.MapDetails, c1 [2]float64) ([][2]float64, []float64) {
	edges := make([][2]float64, 0)
	dist := make([]float64, 0)

	for key, _ := range graph.Nodes {
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

func (graph *DLiteGraph) UpdatePathfinding(mp *maps.MapDetails, newLocation [2]float64) {

	graph.Visited = make(map[[2]float64]bool)

	graph.Km += maps.EuclideanDistance2(graph.Last, newLocation)
	graph.Last = newLocation

	_, exists := graph.TemporaryLocations[newLocation]
	if !exists {
		graph.AddNode(newLocation)
		newEdges, newEdgesCost := graph.GetEdges(mp, newLocation)
		for i := 0; i < len(newEdges); i++ {
			graph.AddEdge(newLocation, newEdges[i], newEdgesCost[i])
		}
		start := time.Now()
		graph.updateVertex(newLocation)
		graph.ComputeShortestPath()
		elapsed := time.Since(start)
		fmt.Println("D* runtime", elapsed)
		graph.ReturnShortestPath()
	}

	// if the old last is in temporary, remove it

}
