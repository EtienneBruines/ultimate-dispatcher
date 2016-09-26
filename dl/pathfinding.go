package dl

import "engo.io/engo"

func SetRoute(from, to engo.Point) Route {
	// Go to node closest to where we wanna go
	dest := CurrentMap.NearestNode(to)

	// Going for an A* algorithm, with Euclidean-distance as heuristic (including the cost)
	h := func(curr, goal, pos *RouteNode) float32 {
		dx := pos.Location.X - goal.Location.X
		dy := pos.Location.Y - goal.Location.Y
		dx2 := pos.Location.X - curr.Location.X
		dy2 := pos.Location.Y - curr.Location.Y
		return dx*dx + dy*dy + (dx2*dx2 + dy2*dy2)
	}

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
				continue // skip whatever we've already visited
			}

			childNode := CurrentMap.Node(connID)
			heuristic := h(curr, dest, nNode)
			queue.Enqueue(queueItem{Route: Route{Nodes: append(n.Route.Nodes, childNode)}}, heuristic)
			visited[connID] = struct{}{}
		}
	}

	if !goalReached {
		panic("No route found")
	}

	return route
}