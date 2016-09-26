package dl

import (
	"engo.io/engo"
	"fmt"

	"github.com/luxengine/math"
)

type IncidentCarSpeeding struct {
	Start        engo.Point
	Goal         engo.Point

	currentRoute Route
	finished     bool

	Location     *engo.Point
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

func (i* IncidentCarSpeeding) SetLocation(loc *engo.Point) {
	i.Location = loc
}

func (i *IncidentCarSpeeding) Update(dt float32) {
	// Compute route if required
	if len(i.currentRoute.Nodes) < 1 {
		i.currentRoute = SetRoute(*i.Location, i.Goal)
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
	const speed = 200
	
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
