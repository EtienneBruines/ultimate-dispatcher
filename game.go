package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EtienneBruines/ultimate-dispatcher/dl"
	"github.com/EtienneBruines/ultimate-dispatcher/ui"
	"github.com/luxengine/math"
)

const (
	title = "Ultimate Dispatcher"

	KeyboardScrollSpeed = 400
	EdgeScrollSpeed     = KeyboardScrollSpeed
	EdgeWidth           = 20
	ZoomSpeed           = -0.125
)

type Game struct {
	Paused bool
}

var TheGame = &Game{}

func (g *Game) Preload() {
	engo.Files.Register(".map", &dl.MapLoader{})

	engo.Files.Load("maps/1.map")
	engo.Files.Load("fonts/Roboto-Regular.ttf")
}

func (g *Game) Setup(w *ecs.World) {
	common.SetBackground(color.NRGBA{100, 100, 100, 255})
	rs := &common.RenderSystem{}
	ms := &common.MouseSystem{}
	iss := &dl.IncidentSystem{}
	ds := &dl.DispatchSystem{}

	w.AddSystem(&common.CameraSystem{})
	w.AddSystem(rs)
	w.AddSystem(ms)
	w.AddSystem(common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	//w.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	w.AddSystem(&common.MouseZoomer{ZoomSpeed})
	w.AddSystem(&dl.IncidentDebugSystem{})
	w.AddSystem(ds)
	w.AddSystem(iss)

	/*
		mResource, err := engo.Files.Resource("maps/1.map")
		if err != nil {
			panic(err)
		}

		m, ok := mResource.(*dl.Map)
		if !ok {
			panic(fmt.Errorf("Map resource is not of type *Map: %s", "maps/1.map"))
		}
	*/

	m := dl.RandomMap(10, 10, 100, 100)
	m.Initialize()

	for _, node := range m.Nodes {
		type mapEntity struct {
			ecs.BasicEntity
			common.RenderComponent
			common.SpaceComponent
		}

		/* We can add it if we want to enable it for debugging
		me := mapEntity{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: nodeGraphic, Color: NodeColor},
			SpaceComponent:  common.SpaceComponent{node.Location, NodeSize, NodeSize, 0},
		}

		rs.Add(&me.BasicEntity, &me.RenderComponent, &me.SpaceComponent)
		*/

		// Render roads - TODO: optimize
		for _, conn := range node.ConnectedTo {
			type roadEntity struct {
				ecs.BasicEntity
				common.RenderComponent
				common.SpaceComponent
			}
			loc, length, rot := ComputeRoad(node.Location, m.Node(conn).Location, ui.RoadSize)

			road := roadEntity{
				BasicEntity:     ecs.NewBasic(),
				RenderComponent: common.RenderComponent{Drawable: ui.RoadGraphic, Color: ui.RoadColor},
				SpaceComponent:  common.SpaceComponent{loc, length, ui.RoadSize, rot},
			}
			rs.Add(&road.BasicEntity, &road.RenderComponent, &road.SpaceComponent)

			/*
				connNode := m.Node(conn)
				var found bool
				for _, n := range connNode.ConnectedTo {
					if n == node.ID {
						found = true
						break
					}
				}
				if !found {
					connNode.ConnectedTo = append(connNode.ConnectedTo, node.ID)
				}
			*/
		}
	}

	// Now let's move on to the "incidents"
	start := engo.Point{100, 100}
	goal := engo.Point{500, 100}
	incidents := []dl.IncidentComponent{
		{Location: &start,
			Incident: &dl.IncidentCarSpeeding{
				Start: start,
				Goal: goal,
			},
			Reports: []dl.IncidentReportComponent{
			{&engo.Point{100, 145}, "IncidentCarAccident", 1, 1, dl.UrgencyNeutral},
			{&engo.Point{130, 104}, "IncidentCarAccident", 1, 1, dl.UrgencyNotUrgent},
			{&engo.Point{165, 102}, "IncidentCarAccident", 1, 2, dl.UrgencyUrgent},
		}},
	}

	for id := range incidents {
		iss.Spawn(incidents[id])
	}

	// Now let's see if we can get some police ready for the incident

	unitTypes, err := dl.LoadPoliceUnits("assets/units/police.yaml")
	if err != nil {
		panic(err)
	}

	unitLocations := []engo.Point{
		{300, 300},
		{500, 500},
		{400, 100},
		{900, 900},
	}
	units := []dl.PoliceComponent{
		{Unit: unitTypes.ByName("Car")},
		{Unit: unitTypes.ByName("Car")},
		{Unit: unitTypes.ByName("Bike Light")},
		{Unit: unitTypes.ByName("Van w/ Cells")},
	}
	for i, unit := range units {
		pe := dl.PoliceEntity{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: ui.PoliceGraphic, Color: ui.PoliceColor, TextureAlignment: common.AlignCenter},
			SpaceComponent:  common.SpaceComponent{unitLocations[i], ui.PoliceSize * unit.Unit.Size, ui.PoliceSize * unit.Unit.Size, 0},
			PoliceComponent: unit,
		}
		pe.SetZIndex(ui.PoliceZIndex)
		pe.PoliceComponent.Location = &pe.SpaceComponent.Position
		rs.Add(&pe.BasicEntity, &pe.RenderComponent, &pe.SpaceComponent)
		ms.Add(&pe.BasicEntity, &pe.MouseComponent, &pe.SpaceComponent, &pe.RenderComponent)
		ds.AddPolice(&pe.BasicEntity, &pe.RenderComponent, &pe.SpaceComponent, &pe.MouseComponent, &pe.PoliceComponent)
	}
}

// ComputeRoad computes data needed to position roads between two points
func ComputeRoad(from, to engo.Point, height float32) (engo.Point, float32, float32) {
	roadLength := math.Sqrt(
		math.Pow(from.X-to.X, 2) +
			math.Pow(from.Y-to.Y, 2),
	)

	a := to.Y - from.Y
	b := roadLength
	c := to.X - from.X
	if c == 0 {
		if a > 0 {
			return from, roadLength, 90
		} else {
			return from, roadLength, -90
		}
	}

	rotation_rad := math.Acos((-math.Pow(a, 2) + math.Pow(b, 2) + math.Pow(c, 2)) / (2 * b * c))
	rotation := 180 * (rotation_rad / math.Pi)

	return from, roadLength, -rotation
}

func (g *Game) Type() string {
	return "GameScene"
}

func main() {
	opts := engo.RunOptions{
		Title:          title,
		StandardInputs: true,
		Height:         860,
		Width:          800,
	}
	engo.Run(opts, TheGame)
}
