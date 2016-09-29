package dl

import (
	"engo.io/engo"
)

func SetRoute(from, to engo.Point, h func(curr, goal, pos *RouteNode) float32) Route {
	// Go to node closest to where we wanna go
	dest := CurrentMap.NearestNode(to)

	// Going for an A* algorithm, with Euclidean-distance as heuristic (including the cost)

	visited := make(map[uint32]struct{})
	curr := CurrentMap.NearestNode(from)

	type queueItem struct {
		Route Route
	}

	var queue PriorityQueue
	queue.Enqueue(queueItem{Route: Route{Nodes: []*RouteNode{curr}}}, 0)

	var goalReached bool
	var route Route

	for !goalReached && len(queue.values) > 0 {
		// Dequeue
		next := queue.Dequeue()
		n := next.(queueItem)
		nNode := n.Route.Nodes[len(n.Route.Nodes)-1]

		if nNode.ID == dest.ID {
			goalReached = true
			route = n.Route
			break
		}

		for _, connID := range nNode.ConnectedTo {
			if _, ok := visited[connID]; ok {
				continue // don't queue whatever we've already queued once
			}

			childNode := CurrentMap.Node(connID)
			heuristic := h(curr, dest, childNode)

			oldRoute := make([]*RouteNode, len(n.Route.Nodes), len(n.Route.Nodes)+1)
			copy(oldRoute, n.Route.Nodes)
			queue.Enqueue(queueItem{Route: Route{Nodes: append(oldRoute, childNode)}}, heuristic)

			visited[connID] = struct{}{}
		}
	}

	if !goalReached {
		panic("No route found")
	}

	return route
}
