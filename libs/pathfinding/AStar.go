package pathfinding

import (
	"container/heap"
	"fd/libs/maps"
	"fmt"
	"math"
	"time"
)

type AStarKey struct {
	Location Location
	F        float64
}

type AStarGraph struct {
	Start Location
	Goal  Location

	Nodes  map[Location]float64 // g-score
	Edges  map[Location]map[Location]float64
	Parent map[Location]*Location

	PriorityQueue *AStarPQ
	Visited       map[Location]bool
}

type AStarPQ struct {
	queue    []AStarKey
	indexMap map[Location]int // Maps node (coordinate) to its index in queue
}

func (pq *AStarPQ) Len() int { return len(pq.queue) }

func (pq *AStarPQ) Less(i, j int) bool {
	if pq.queue[i].F < pq.queue[j].F {
		return true
	}
	return false
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

func (pq *AStarPQ) topKey() AStarKey {
	if len(pq.queue) == 0 {
		return AStarKey{F: math.Inf(1)}
	}
	return pq.queue[0] // The top of the heap is always at index 0
}

func (pq *AStarPQ) Remove(location Location) {
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

func (pq *AStarPQ) Update(location Location, fCost float64) {
	index, exists := pq.indexMap[location]
	if !exists {
		heap.Push(pq, AStarKey{Location: location, F: fCost})
		return
	}

	pq.queue[index].F = fCost
	heap.Fix(pq, index) // Restore heap property, swap will update index mappings for us
}

func AStarInit(mp *maps.MapDetails, start Location, goal Location) *AStarGraph {

	masterAStar := &AStarGraph{
		Start: start,
		Goal:  goal,

		Nodes:  make(map[Location]float64),
		Edges:  make(map[Location]map[Location]float64),
		Parent: make(map[Location]*Location),

		PriorityQueue: &AStarPQ{
			queue:    make([]AStarKey, 0),
			indexMap: make(map[Location]int),
		},
		Visited: make(map[Location]bool),
	}

	for i := 0; i < len(mp.Intersections); i++ {

		interestLocation := Location{X: mp.Intersections[i][0], Y: mp.Intersections[i][1]}
		masterAStar.Nodes[interestLocation] = math.Inf(1)

		indivMap := make(map[Location]float64)
		for j := 0; j < len(mp.Intersections_edges_int[i]); j++ {
			edgeArray := mp.Intersections[mp.Intersections_edges_int[i][j]]
			edgeLocation := Location{edgeArray[0], edgeArray[1]}
			indivMap[edgeLocation] = mp.Intersections_edgelength[i][j]
		}
		masterAStar.Edges[interestLocation] = indivMap

		masterAStar.Parent[interestLocation] = nil
	}

	return masterAStar
}

func (graph *AStarGraph) CalculateKey(location Location) AStarKey {
	F := graph.Nodes[location] + maps.EuclideanDistance(location.X, location.Y, graph.Goal.X, graph.Goal.Y)
	return AStarKey{Location: location, F: F}
}

func (graph *AStarGraph) AddNode(location Location) {

	if _, exists := graph.Nodes[location]; !exists {
		graph.Nodes[location] = math.Inf(1)
		graph.Edges[location] = make(map[Location]float64)
	}

}

func (graph *AStarGraph) AddEdge(from Location, to Location, cost float64) {

	if _, exists := graph.Nodes[from]; !exists {
		graph.Nodes[from] = math.Inf(1)
	}

	if _, exists := graph.Nodes[to]; !exists {
		graph.Nodes[to] = math.Inf(1)
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

func (graph *AStarGraph) ComputeShortestPath() [][]float64 {

	count := 0

	start := time.Now()

	for graph.PriorityQueue.Len() > 0 {

		current := heap.Pop(graph.PriorityQueue).(AStarKey)
		//fmt.Println("current", current.Location, current.F)

		if current.Location == graph.Goal {

			elapsed := time.Since(start)
			fmt.Println("A* Runtime", elapsed)
			fmt.Println("A* count", count)

			ans := [][]float64{}

			currentPath := &graph.Goal

			for currentPath != nil {

				ans = append(ans, []float64{currentPath.X, currentPath.Y})
				currentPath = graph.Parent[*currentPath]

			}

			ans2 := [][]float64{}
			for i := len(ans) - 1; i >= 0; i-- {
				ans2 = append(ans2, ans[i])
			}

			fmt.Println("ASTAR ANS", ans2)

			sum := 0.0
			for i := 0; i < len(ans)-1; i++ {
				sum += maps.EuclideanDistance(ans[i][0], ans[i][1], ans[i+1][0], ans[i+1][1])
			}
			fmt.Println("sum", sum)

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
					graph.Parent[neighbor] = &Location{X: current.Location.X, Y: current.Location.Y}
					graph.Nodes[neighbor] = tentativeG
					newF := graph.CalculateKey(neighbor)
					//fmt.Println("n", newF, graph.Nodes[neighbor], tentativeG)
					graph.PriorityQueue.Update(neighbor, newF.F) // I thought you would update only if G val is better

				}

			}

		}

	}

	return [][]float64{}

}
