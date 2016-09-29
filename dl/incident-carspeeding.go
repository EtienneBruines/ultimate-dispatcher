package dl

import (
	"engo.io/engo"
	"fmt"

	"github.com/luxengine/math"
)

type IncidentCarSpeeding struct {
	Start engo.Point
	Goal  engo.Point

	currentRoute Route
	finished     bool

	Location *engo.Point
}

var (
	learningRate float32 = 0.5

	criminalMind = neuralNetwork{
		{
			// Distance left
			func(curr, goal, pos *RouteNode) float32 {
				dx := goal.Location.X - pos.Location.X
				dy := goal.Location.Y - pos.Location.Y
				return dx*dx + dy*dy
			},
			-0.5,
		},
		{
			// Distance travelled
			func(curr, goal, pos *RouteNode) float32 {
				dx := pos.Location.X - curr.Location.X
				dy := pos.Location.Y - curr.Location.Y
				return dx*dx + dy*dy
			},
			0.5,
		},
		{
			// Close to a cop
			func(curr, goal, pos *RouteNode) float32 {
				d := float32(math.MaxFloat32)
				for _, cop := range police {
					dx := pos.Location.X - cop.Location.X
					dy := pos.Location.Y - cop.Location.Y
					if d2 := dx*dx + dy*dy; d2 < d {
						d = d2
					}
				}

				if d == 0 {
					d = 0.00001
				}
				return 1 / math.Tanh(d)
			},
			1.0,
		},
	}
)

type neuralNetwork []neuralElement

type neuralElement struct {
	Value  func(curr, goal, pos *RouteNode) float32
	Weight float32
}

func (n neuralNetwork) Value(curr, goal, pos *RouteNode) (total float32) {
	for _, e := range n {
		total += e.Weight * e.Value(curr, goal, pos)
	}
	fmt.Println("value:", total)
	return
}

func (IncidentCarSpeeding) Type() string {
	return "IncidentCarSpeeding"
}

func (IncidentCarSpeeding) Penalty() int {
	return 10
}

func (IncidentCarSpeeding) Reward() int {
	return 100
}

func (i *IncidentCarSpeeding) SetLocation(loc *engo.Point) {
	i.Location = loc
}

func (i *IncidentCarSpeeding) Update(dt float32) {
	// Compute route if required
	if len(i.currentRoute.Nodes) < 1 {
		i.currentRoute = SetRoute(*i.Location, i.Goal, func(curr, goal, pos *RouteNode) float32 {
			return criminalMind.Value(curr, goal, pos)
		})
	}
	i.Move(dt)
}

func (i IncidentCarSpeeding) Resolved() int {
	if i.finished {
		return -1
	}
	return 0
}

func (i *IncidentCarSpeeding) Move(dt float32) {
	const speed = 250

	var distance = speed / 3.6 * dt

	target := i.currentRoute.Nodes[0].Location

	dx := target.X - i.Location.X
	dy := target.Y - i.Location.Y
	dDiagonal := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

	var movementX, movementY float32
	if dDiagonal > distance {
		ratio := dDiagonal / distance
		movementX = dx / ratio
		movementY = dy / ratio
	} else {
		movementX = dx
		movementY = dy
		i.currentRoute.Nodes = i.currentRoute.Nodes[1:]
		if len(i.currentRoute.Nodes) == 0 {
			i.finished = true
		} else {
			fmt.Println(len(i.currentRoute.Nodes))
		}
	}

	i.Location.X += movementX
	i.Location.Y += movementY
}
