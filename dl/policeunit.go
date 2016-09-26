package dl

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"engo.io/engo"
	"github.com/luxengine/math"
	"gopkg.in/yaml.v2"
)

type PoliceCommand uint8

const (
	CommandHold PoliceCommand = iota
	CommandMove
	CommandLookout
	CommandSearchArea
	CommandTrafficControl
)

type PoliceUnitType struct {
	Name             string
	Speed            float32
	Size             float32 `yaml:"size"`
	PassengersPolice int     `yaml:"passenges"`
	PassengersCuffed int     `yaml:"arrested"`
	PassengersTotal  int     `yaml:"total"`
	ViewDistance     float32 `yaml:"distance_view"`
}

type PoliceUnitTypes []PoliceUnitType

func (p PoliceUnitTypes) ByName(name string) PoliceUnitType {
	for id, police := range p {
		if police.Name == name {
			return p[id]
		}
	}
	return PoliceUnitType{}
}

type PoliceComponent struct {
	Location *engo.Point
	Unit     PoliceUnitType

	// Commands stuff
	Commands []PoliceCommand
	Targets  []engo.Point

	CurrentCommand PoliceCommand
	CurrentTarget  engo.Point
	CurrentResolve DispatchSystemIncidentEntity

	// Move-specific info
	CurrentRoute Route
}

func LoadPoliceUnits(filename string) (PoliceUnitTypes, error) {
	ext := filepath.Ext(filename)
	var unmarshal func([]byte, interface{}) error

	switch ext {
	case ".yaml":
		unmarshal = yaml.Unmarshal
	default:
		// Ignore
		return nil, nil
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var units struct {
		Units []PoliceUnitType
	}

	err = unmarshal(b, &units)
	if err != nil {
		return nil, err
	}

	return units.Units, nil
}

func (p *PoliceComponent) QueueCommand(c PoliceCommand, target engo.Point) {
	p.Commands = append(p.Commands, c)
	p.Targets = append(p.Targets, target)
}

func (p *PoliceComponent) processCommand() (PoliceCommand, engo.Point) {
	if len(p.Commands) == 0 {
		return CommandHold, engo.Point{}
	}

	cmd := p.Commands[0]
	p.Commands = p.Commands[1:]
	target := p.Targets[0]
	p.Targets = p.Targets[1:]
	return cmd, target
}

func (p *PoliceComponent) Update(dt float32) {

}

// Move allows the unit to move to the set destination, at the speed of the update
func (p *PoliceComponent) Move(dt float32) {
	var distance = p.Unit.Speed / 3.6 * dt

	target := p.CurrentRoute.Nodes[0].Location

	dx := target.X - p.Location.X
	dy := target.Y - p.Location.Y
	dDiagonal := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

	var movementX, movementY float32
	if dDiagonal > distance {
		ratio := dDiagonal / distance
		movementX = dx / ratio
		movementY = dy / ratio
	} else {
		movementX = dx
		movementY = dy
		p.CurrentRoute.Nodes = p.CurrentRoute.Nodes[1:]
		if len(p.CurrentRoute.Nodes) == 0 {
			p.CurrentCommand = CommandHold
		} else {
			fmt.Println(len(p.CurrentRoute.Nodes))
		}
	}

	p.Location.X += movementX
	p.Location.Y += movementY
}

func (p *PoliceComponent) Wander(dt float32, location engo.Point) {
	fmt.Println("TODO: wander behavior")
}
