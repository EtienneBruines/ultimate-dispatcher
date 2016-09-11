package main

import (
	"fmt"
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
	IncidentCount int
	Paused        bool

	hoverings map[uint64]bool
}

func (g *Game) StartHovering(uid uint64) {
	if len(g.hoverings) == 0 {
		engo.SetCursor(engo.CursorHand)
	}
	g.hoverings[uid] = true
}

func (g *Game) StopHovering(uid uint64) {
	delete(g.hoverings, uid)
	if len(g.hoverings) == 0 {
		engo.SetCursor(engo.CursorNone)
	}
}

var TheGame = &Game{hoverings: make(map[uint64]bool)}

func (g *Game) Preload() {
	engo.Files.Register(".map", &dl.MapLoader{})

	engo.Files.Load("maps/1.map")
	engo.Files.Load("fonts/Roboto-Regular.ttf")
}

const (
	NodeSize     = 10
	RoadSize     = 2
	IncidentSize = 2 * NodeSize
	policeSize   = 2 * NodeSize
	waypointSize = 1 * NodeSize

	tooltipLineHeight float32 = 24

	closeButton = "close"
)

var (
	NodeColor           = color.NRGBA{0, 255, 255, 255}
	RoadColor           = color.White
	IncidentColor       = color.NRGBA{255, 0, 0, 128}
	IncidentColorHover  = color.NRGBA{153, 0, 0, 255}
	policeColor         = color.NRGBA{0, 0, 255, 128}
	policeColorSelected = color.NRGBA{255, 0, 255, 255}
	policeColorHover    = color.NRGBA{0, 0, 255, 255}
	tooltipColor        = color.NRGBA{230, 230, 230, 240}
	tooltipColorHover   = color.NRGBA{230, 230, 180, 255}
	tooltipColorBorder  = color.Black
	waypointColor       = color.NRGBA{0, 255, 0, 150}

	nodeGraphic     = common.Circle{}
	roadGraphic     = common.Rectangle{}
	incidentGraphic = common.Circle{}
	policeGraphic   = common.Circle{}
	tooltipGraphic  = common.Rectangle{BorderWidth: 1, BorderColor: tooltipColorBorder}
	waypointGraphic = common.Rectangle{}
)

func (g *Game) Setup(w *ecs.World) {
	common.SetBackground(color.NRGBA{100, 100, 100, 255})
	rs := &common.RenderSystem{}
	ms := &common.MouseSystem{}
	ds := &DispatchSystem{}
	ids := &IncidentDetailSystem{}

	w.AddSystem(&common.CameraSystem{})
	w.AddSystem(rs)
	w.AddSystem(ms)
	w.AddSystem(common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	//w.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	w.AddSystem(&common.MouseZoomer{ZoomSpeed})
	w.AddSystem(ds)
	w.AddSystem(ids)

	engo.Input.RegisterButton(closeButton, engo.Escape)

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

		me := mapEntity{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: nodeGraphic, Color: NodeColor},
			SpaceComponent:  common.SpaceComponent{node.Location, NodeSize, NodeSize, 0},
		}
		rs.Add(&me.BasicEntity, &me.RenderComponent, &me.SpaceComponent)

		// Render roads - TODO: optimize
		for _, conn := range node.ConnectedTo {
			type roadEntity struct {
				ecs.BasicEntity
				common.RenderComponent
				common.SpaceComponent
			}
			loc, length, rot := ComputeRoad(node.Location, m.Node(conn).Location, RoadSize)

			road := roadEntity{
				BasicEntity:     ecs.NewBasic(),
				RenderComponent: common.RenderComponent{Drawable: roadGraphic, Color: RoadColor},
				SpaceComponent:  common.SpaceComponent{loc, length, RoadSize, rot},
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
	incidents := []Incident{
		{Type: IncidentCarSpeeding, Location: engo.Point{100, 100}, Reports: []IncidentReport{
			{IncidentCarSpeeding, 1, 1, UrgencyNeutral},
			{IncidentCarSpeeding, 1, 1, UrgencyNotUrgent},
			{IncidentCarAccident, 1, 2, UrgencyUrgent},
		}},
	}

	for _, in := range incidents {
		// Look for nearest node, to prevent issues
		loc := m.NearestNode(in.Location)

		type IncidentEntity struct {
			ecs.BasicEntity
			common.RenderComponent
			common.SpaceComponent
			common.MouseComponent
			IncidentComponent
		}

		ie := IncidentEntity{
			BasicEntity:       ecs.NewBasic(),
			RenderComponent:   common.RenderComponent{Drawable: incidentGraphic, Color: IncidentColor},
			SpaceComponent:    common.SpaceComponent{loc.Location, IncidentSize, IncidentSize, 0},
			IncidentComponent: IncidentComponent{in},
		}
		ie.RenderComponent.SetZIndex(5)
		rs.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent)
		ms.Add(&ie.BasicEntity, &ie.MouseComponent, &ie.SpaceComponent, &ie.RenderComponent)
		ids.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.MouseComponent, &ie.IncidentComponent)
		ds.AddIncident(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent, &ie.MouseComponent, &ie.IncidentComponent)
		g.IncidentCount++
	}

	// Show the incident counter in the corner
	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: 12,
	}
	if err := fnt.CreatePreloaded(); err != nil {
		panic(err)
	}

	incidentLabel := ui.Label{
		BasicEntity:     ecs.NewBasic(),
		Font:            fnt,
		SpaceComponent:  common.SpaceComponent{engo.Point{0, 0}, 100, 20, 0},
		RenderComponent: common.RenderComponent{Scale: engo.Point{0.5, 0.5}},
	}
	incidentLabel.SetText(fmt.Sprintf("Active Incidents: %d", g.IncidentCount))
	incidentLabel.RenderComponent.SetShader(common.HUDShader)

	rs.Add(&incidentLabel.BasicEntity, &incidentLabel.RenderComponent, &incidentLabel.SpaceComponent)

	// Now let's see if we can get some police ready for the incident

	units := []dl.Police{
		{ID: 1},
	}
	for _, unit := range units {
		pe := PoliceEntity{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: policeGraphic, Color: policeColor},
			SpaceComponent:  common.SpaceComponent{engo.Point{300, 300}, policeSize, policeSize, 0},
			PoliceComponent: dl.PoliceComponent{unit},
		}
		pe.PoliceComponent.Police.Location = &pe.SpaceComponent.Position
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

type PoliceEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	dl.PoliceComponent
}

type DispatchSystemPoliceEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
	*common.MouseComponent
	*dl.PoliceComponent
}

type DispatchSystemIncidentEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
	*common.MouseComponent
	*IncidentComponent
}

type DispatchSystem struct {
	police    map[uint64]DispatchSystemPoliceEntity
	incidents map[uint64]DispatchSystemIncidentEntity

	active            uint64
	submenuTarget     engo.Point
	submenuActive     bool
	submenuBackground ui.Graphic
	submenuActions    []*ui.Button
	mouseTracker      common.MouseComponent
	wpEntity          ui.Button
}

func (d *DispatchSystem) QueueCommand(c dl.PoliceCommand) {
	unit := d.police[d.active]
	// create a temporary node for the submenuTarget

	nearest := dl.CurrentMap.NearestNode(d.submenuTarget)
	if nearest.Temporary {
		nearest.TemporaryUsers++
	} else {
		temp := new(dl.RouteNode)
		temp.Location = d.submenuTarget
		temp.ID = dl.NewMapID()
		temp.Temporary = true
		temp.TemporaryUsers = 1

		// And also add the second connected City
		minDistance := float32(math.MaxFloat32)
		var secondNearest *dl.RouteNode
		for _, connection := range nearest.ConnectedTo {
			conn := dl.CurrentMap.Node(connection)
			if d := conn.Location.PointDistance(d.submenuTarget); d < minDistance {
				minDistance = d
				secondNearest = conn
			}
		}

		nearest.ConnectedTo = append(nearest.ConnectedTo, temp.ID)
		secondNearest.ConnectedTo = append(secondNearest.ConnectedTo, temp.ID)
		temp.ConnectedTo = []uint32{nearest.ID, secondNearest.ID}

		dl.CurrentMap.AddNode(temp)
		// TODO: clean this up later to prevent (relatively slow) memory leaking
	}

	unit.Police.QueueCommand(c, d.submenuTarget)
}

func (d *DispatchSystem) New(w *ecs.World) {
	d.police = make(map[uint64]DispatchSystemPoliceEntity)
	d.incidents = make(map[uint64]DispatchSystemIncidentEntity)

	d.mouseTracker.Track = true
	mouseTrackerBasic := ecs.NewBasic()

	actions := []struct {
		Name    string
		OnClick func(*ui.Button)
	}{
		{Name: "Search area", OnClick: func(*ui.Button) {
			d.QueueCommand(dl.CommandMove)
			d.QueueCommand(dl.CommandSearchArea)
		}},
		{Name: "Hold watch", OnClick: func(*ui.Button) {
			d.QueueCommand(dl.CommandMove)
			d.QueueCommand(dl.CommandLookout)
		}},
	}

	d.submenuBackground = ui.Graphic{
		BasicEntity: ecs.NewBasic(),
		RenderComponent: common.RenderComponent{
			Drawable: tooltipGraphic,
			Color:    tooltipColor,
		},
		SpaceComponent: common.SpaceComponent{
			Width:  200,
			Height: tooltipLineHeight * float32(len(actions)),
		},
	}
	d.submenuBackground.RenderComponent.SetShader(common.HUDShader)

	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: float64(tooltipLineHeight),
	}
	if err := fnt.CreatePreloaded(); err != nil {
		panic(err)
	}

	for _, action := range actions {
		but := ui.NewButton(fnt, action.Name)
		but.OnClick = func(b *ui.Button) {
			b.OnMouseOut(b) // TODO: verify if we need this?
			action.OnClick(b)
		}
		but.OnMouseOver = func(b *ui.Button) {
			b.Graphic.Color = tooltipColorHover
			TheGame.StartHovering(but.Graphic.ID())
		}
		but.OnMouseOut = func(b *ui.Button) {
			b.Graphic.Color = tooltipColor
			TheGame.StopHovering(but.Graphic.ID())
		}
		but.Label.Width = 200
		but.Label.Height = tooltipLineHeight
		but.Label.SetZIndex(10)
		but.Label.SetShader(common.TextHUDShader)
		but.Graphic.Color = tooltipColor
		but.Graphic.Drawable = tooltipGraphic
		but.Graphic.Width = 200
		but.Graphic.Height = tooltipLineHeight
		but.Graphic.SetZIndex(9)
		but.Graphic.RenderComponent.SetShader(common.HUDShader)
		d.submenuActions = append(d.submenuActions, but)
	}

	d.wpEntity = ui.Button{
		Graphic: ui.Graphic{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: waypointGraphic, Color: waypointColor, Hidden: true},
			SpaceComponent:  common.SpaceComponent{Width: waypointSize, Height: waypointSize},
		},
	}
	d.wpEntity.Graphic.SetZIndex(5)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&d.submenuBackground.BasicEntity, &d.submenuBackground.RenderComponent, &d.submenuBackground.SpaceComponent)
			for _, sa := range d.submenuActions {
				sys.Add(&sa.Label.BasicEntity, &sa.Label.RenderComponent, &sa.Label.SpaceComponent)
				sys.Add(&sa.Graphic.BasicEntity, &sa.Graphic.RenderComponent, &sa.Graphic.SpaceComponent)
			}
			sys.Add(&d.wpEntity.Graphic.BasicEntity, &d.wpEntity.Graphic.RenderComponent, &d.wpEntity.Graphic.SpaceComponent)
		case *common.MouseSystem:
			for _, sa := range d.submenuActions {
				sys.Add(&sa.Graphic.BasicEntity, &sa.MouseComponent, &sa.Graphic.SpaceComponent, &sa.Graphic.RenderComponent)
			}
			sys.Add(&mouseTrackerBasic, &d.mouseTracker, nil, nil)
			sys.Add(&d.wpEntity.Graphic.BasicEntity, &d.wpEntity.MouseComponent, &d.wpEntity.Graphic.SpaceComponent, &d.wpEntity.Graphic.RenderComponent)
		}
	}

	d.hideSubmenu()
}

func (d *DispatchSystem) hideSubmenu() {
	d.submenuActive = false
	d.submenuBackground.Hidden = true
	for _, action := range d.submenuActions {
		action.Label.Hidden = true
		action.Graphic.Hidden = true
		action.Graphic.SpaceComponent.Position = engo.Point{-math.MaxFloat32, -math.MaxFloat32}
		TheGame.StopHovering(action.Graphic.ID())
	}
}

func (d *DispatchSystem) showSubmenu(pos engo.Point) {
	d.submenuTarget = pos
	d.submenuActive = true
	d.submenuBackground.Hidden = false
	d.submenuBackground.Position = pos
	var offset float32
	for _, action := range d.submenuActions {
		action.Label.Position.X = pos.X
		action.Label.Position.Y = pos.Y + offset
		action.Label.Hidden = false

		action.Graphic.Position.X = pos.X
		action.Graphic.Position.Y = pos.Y + offset
		action.Graphic.Hidden = false

		offset += tooltipLineHeight
	}
}

func (d *DispatchSystem) AddPolice(b *ecs.BasicEntity, r *common.RenderComponent, s *common.SpaceComponent, m *common.MouseComponent, p *dl.PoliceComponent) {
	d.police[b.ID()] = DispatchSystemPoliceEntity{b, r, s, m, p}
}

func (d *DispatchSystem) AddIncident(b *ecs.BasicEntity, r *common.RenderComponent, s *common.SpaceComponent, m *common.MouseComponent, i *IncidentComponent) {
	d.incidents[b.ID()] = DispatchSystemIncidentEntity{b, r, s, m, i}
}

func (d *DispatchSystem) Remove(b ecs.BasicEntity) {
	delete(d.police, b.ID())
	delete(d.incidents, b.ID())
}

func (d *DispatchSystem) Update(dt float32) {
	// Allow us to select a police unit
	if d.active == 0 {
		for id, police := range d.police {
			if police.MouseComponent.Enter {
				police.Color = policeColorHover
				TheGame.StartHovering(id)
			} else if police.MouseComponent.Leave {
				police.Color = policeColor
				TheGame.StopHovering(id)
			}
			if police.MouseComponent.Clicked {
				police.Color = policeColorSelected
				d.active = id
				d.wpEntity.Graphic.Hidden = false
				TheGame.StopHovering(id)
				return
			}
		}
	}

	// If we've selected a police unit, we can issue commands
	if d.active > 0 {
		police := d.police[d.active]

		if !d.submenuActive {

			// We can issue commands anywhere we want, as long as it's connected to roads.
			mX, mY := d.mouseTracker.MouseX, d.mouseTracker.MouseY
			mP := engo.Point{mX, mY}
			// Check which city is closest, and try to snap to that road
			nearest := dl.CurrentMap.NearestNode(mP)
			// Now figure out which of the roads to snap to
			// Source for this "distance" method, https://stackoverflow.com/a/6853926/3243814
			distanceFunc := func(point, l1, l2 engo.Point) float32 {
				A, B := point.X-l1.X, point.Y-l1.Y
				C, D := l2.X-l1.X, l2.Y-l1.Y
				dot := A*C + B*D
				len_sq := math.Pow(C, 2) + math.Pow(D, 2)
				param := float32(-1)
				if len_sq != 0 {
					param = dot / len_sq
				}
				var xx, yy float32
				if param < 0 {
					xx, yy = l1.X, l1.Y
				} else if param > 1 {
					xx, yy = l2.X, l2.Y
				} else {
					xx, yy = l1.X+param*C, l1.Y+param*D
				}
				dx, dy := point.X-xx, point.Y-yy
				return math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
			}

			minDistance := float32(math.MaxFloat32)
			var secondNearest *dl.RouteNode
			for _, connected := range nearest.ConnectedTo {
				conn := dl.CurrentMap.Node(connected)
				if d := distanceFunc(mP, nearest.Location, conn.Location); d < minDistance {
					minDistance = d
					secondNearest = conn
				}
			}

			// Then figure out where on the road we want to snap
			// distance on road² = distance to node² - distance to road²
			dNode := nearest.Location.PointDistance(mP)
			dRoad := minDistance
			dOnRoad := math.Sqrt(math.Pow(dNode, 2) - math.Pow(dRoad, 2))

			// Now we have to use triangle similarity to figure out where to place it
			dX, dY := nearest.Location.X-secondNearest.Location.X, nearest.Location.Y-secondNearest.Location.Y
			dDiag := math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2))
			ratio := dDiag / dOnRoad // note: dDiag should always be larger
			waypointdx, waypointdy := dX/ratio, dY/ratio
			waypoint := engo.Point{
				X: nearest.Location.X - waypointdx - waypointSize/2,
				Y: nearest.Location.Y - waypointdy - waypointSize/2,
			}

			// If we've snapped, we should create some kind of "waypoint-icon" player can click
			d.wpEntity.Graphic.SpaceComponent.Position = waypoint

			// Player can click, and will open submenu
			if d.wpEntity.MouseComponent.Clicked {
				d.showSubmenu(waypoint)
			}
		}

		// Check if we're using the submenu, to allow command issuing
		var submenuUsed bool
		if d.submenuActive {
			for _, action := range d.submenuActions {
				if action.Clicked {
					action.OnClick(action)
					submenuUsed = true
					break
				}
				if action.Enter {
					action.OnMouseOver(action)
				} else if action.Leave {
					action.OnMouseOut(action)
				}
			}
		}

		// Allow for cancel behavior
		if engo.Input.Button(closeButton).JustPressed() || police.MouseComponent.Clicked || submenuUsed {
			police.Color = policeColor
			d.active = 0
			TheGame.StopHovering(police.ID())
			d.wpEntity.Graphic.Hidden = true
			if d.submenuActive {
				d.hideSubmenu()
			}
		}
	}

	// Process all commands given to any units
	for _, p := range d.police {
		p.Police.Update(dt)
	}
}

type IncidentType uint8

const (
	IncidentCarAccident IncidentType = iota
	IncidentCarSpeeding
	IncidentHomeRobbery
	IncidentPublicIntoxication
)

type UrgencyLevel uint8

const (
	UrgencyCritical UrgencyLevel = iota
	UrgencyUrgent
	UrgencyNeutral
	UrgencyNotUrgent
)

type Incident struct {
	Location engo.Point
	Award    int
	Penalty  int
	Type     IncidentType

	Reports []IncidentReport
}

type IncidentReport struct {
	Type      IncidentType
	MinAmount uint8
	MaxAmount uint8
	Urgency   UrgencyLevel
}

type IncidentComponent struct {
	Incident Incident
}

type IncidentDetailSystemEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.MouseComponent
	*IncidentComponent
}

type IncidentDetailSystem struct {
	incidents map[uint64]IncidentDetailSystemEntity

	world  *ecs.World
	active uint64
}

func (d *IncidentDetailSystem) New(w *ecs.World) {
	d.world = w
	d.incidents = make(map[uint64]IncidentDetailSystemEntity)
}

func (d *IncidentDetailSystem) Add(b *ecs.BasicEntity, r *common.RenderComponent, m *common.MouseComponent, i *IncidentComponent) {
	d.incidents[b.ID()] = IncidentDetailSystemEntity{b, r, m, i}
}

func (d *IncidentDetailSystem) Remove(b ecs.BasicEntity) {
	delete(d.incidents, b.ID())
}

func (d *IncidentDetailSystem) Update(dt float32) {
	/*
		//if d.active == 0 {
		for uid, in := range d.incidents {
			if in.MouseComponent.Enter {
				in.Color = IncidentColorHover
				engo.SetCursor(engo.CursorHand)
				d.active = uid
			} else if in.MouseComponent.Leave {
				in.Color = IncidentColor
				engo.SetCursor(engo.CursorNone)
				d.active = 0
			}
		}
		//}

		if d.active > 0 {

		}
	*/
}

type HUD struct {
}

func (h *HUD) IncidentDetail(ic *IncidentComponent) {
	/*
		// TODO: this was moved to a later stage, not the PoC
		maxReports := 5

		reports := ic.Incident.Reports
		if len(reports) > maxReports {
			reports = reports[len(reports)-5:]
		}

		width := engo.WindowWidth() / len(reports)
		for _, report := range reports {

		}
	*/
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
