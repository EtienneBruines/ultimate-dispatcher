package dl

import (
	"image/color"
	"log"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EtienneBruines/ultimate-dispatcher/ui"
	"github.com/luxengine/math"
)

const (
	closeButton = "close"
)

type PoliceEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	PoliceComponent
}

type DispatchSystemPoliceEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
	*common.MouseComponent
	*PoliceComponent
}

type DispatchSystemIncidentEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
	*common.MouseComponent
	*IncidentComponent
}

type DispatchSystemIncidentReportEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
	*common.MouseComponent
	*IncidentReportComponent
}

var (
	police          map[uint64]DispatchSystemPoliceEntity
	incidents       map[uint64]DispatchSystemIncidentEntity
	incidentReports map[uint64]DispatchSystemIncidentReportEntity
)

type DispatchSystem struct {
	active            uint64
	submenuTarget     engo.Point
	submenuActive     bool
	submenuBackground ui.Graphic
	submenuActions    []*ui.Button
	mouseTracker      common.MouseComponent
	wpEntity          ui.Button
}

func (d *DispatchSystem) QueueCommand(c PoliceCommand) {
	unit := police[d.active]
	// create a temporary node for the submenuTarget

	nearest := CurrentMap.NearestNode(d.submenuTarget)
	if nearest.Temporary {
		nearest.TemporaryUsers++
	} else {
		temp := new(RouteNode)
		temp.Location = d.submenuTarget
		temp.ID = NewMapID()
		temp.Temporary = true
		temp.TemporaryUsers = 1

		// And also add the second connected City
		minDistance := float32(math.MaxFloat32)
		var secondNearest *RouteNode
		for _, connection := range nearest.ConnectedTo {
			conn := CurrentMap.Node(connection)
			if d := conn.Location.PointDistance(d.submenuTarget); d < minDistance {
				minDistance = d
				secondNearest = conn
			}
		}

		nearest.ConnectedTo = append(nearest.ConnectedTo, temp.ID)
		secondNearest.ConnectedTo = append(secondNearest.ConnectedTo, temp.ID)
		temp.ConnectedTo = []uint32{nearest.ID, secondNearest.ID}

		CurrentMap.AddNode(temp)
		// TODO: clean this up later to prevent (relatively slow) memory leaking
	}

	unit.QueueCommand(c, d.submenuTarget)
}

func (d *DispatchSystem) New(w *ecs.World) {
	police = make(map[uint64]DispatchSystemPoliceEntity)
	incidents = make(map[uint64]DispatchSystemIncidentEntity)

	engo.Input.RegisterButton(closeButton, engo.Escape)

	d.mouseTracker.Track = true
	mouseTrackerBasic := ecs.NewBasic()

	actions := []struct {
		Name    string
		OnClick func(*ui.Button)
	}{
		{Name: "Search area", OnClick: func(*ui.Button) {
			d.QueueCommand(CommandMove)
			d.QueueCommand(CommandSearchArea)
		}},
		{Name: "Hold watch", OnClick: func(*ui.Button) {
			d.QueueCommand(CommandMove)
			d.QueueCommand(CommandLookout)
		}},
	}

	d.submenuBackground = ui.Graphic{
		BasicEntity: ecs.NewBasic(),
		RenderComponent: common.RenderComponent{
			Drawable: ui.TooltipGraphic,
			Color:    ui.TooltipColor,
		},
		SpaceComponent: common.SpaceComponent{
			Width:  200,
			Height: ui.TooltipLineHeight * float32(len(actions)),
		},
	}
	d.submenuBackground.RenderComponent.SetShader(common.HUDShader)

	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: float64(ui.TooltipLineHeight),
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
			b.Graphic.Color = ui.TooltipColorHover
			ui.StartHovering(but.Graphic.ID())
		}
		but.OnMouseOut = func(b *ui.Button) {
			b.Graphic.Color = ui.TooltipColor
			ui.StopHovering(but.Graphic.ID())
		}
		but.Label.Width = 200
		but.Label.Height = ui.TooltipLineHeight
		but.Label.SetZIndex(10)
		but.Label.SetShader(common.TextHUDShader)
		but.Graphic.Color = ui.TooltipColor
		but.Graphic.Drawable = ui.TooltipGraphic
		but.Graphic.Width = 200
		but.Graphic.Height = ui.TooltipLineHeight
		but.Graphic.SetZIndex(9)
		but.Graphic.RenderComponent.SetShader(common.HUDShader)
		d.submenuActions = append(d.submenuActions, but)
	}

	d.wpEntity = ui.Button{
		Graphic: ui.Graphic{
			BasicEntity:     ecs.NewBasic(),
			RenderComponent: common.RenderComponent{Drawable: ui.WaypointGraphic, Color: ui.WaypointColor, Hidden: true},
			SpaceComponent:  common.SpaceComponent{Width: ui.WaypointSize, Height: ui.WaypointSize},
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
		ui.StopHovering(action.Graphic.ID())
	}
}

func (d *DispatchSystem) showSubmenu(pos engo.Point) {
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

		offset += ui.TooltipLineHeight
	}
}

func (d *DispatchSystem) AddPolice(b *ecs.BasicEntity, r *common.RenderComponent, s *common.SpaceComponent, m *common.MouseComponent, p *PoliceComponent) {
	police[b.ID()] = DispatchSystemPoliceEntity{b, r, s, m, p}
}

func (d *DispatchSystem) AddIncident(b *ecs.BasicEntity, r *common.RenderComponent, s *common.SpaceComponent, m *common.MouseComponent, i *IncidentComponent) {
	incidents[b.ID()] = DispatchSystemIncidentEntity{b, r, s, m, i}
}

func (d *DispatchSystem) AddIncidentReport(b *ecs.BasicEntity, r *common.RenderComponent, s *common.SpaceComponent, m *common.MouseComponent, i *IncidentReportComponent) {
	incidentReports[b.ID()] = DispatchSystemIncidentReportEntity{b, r, s, m, i}
}

func (d *DispatchSystem) Remove(b ecs.BasicEntity) {
	delete(police, b.ID())
	delete(incidents, b.ID())
	delete(incidentReports, b.ID())
}

func (d *DispatchSystem) Update(dt float32) {
	// Allow us to select a police unit
	if d.active == 0 {
		for id, police := range police {
			if police.MouseComponent.Enter {
				police.Color = ui.PoliceColorHover
				ui.StartHovering(id)
			} else if police.MouseComponent.Leave {
				police.Color = ui.PoliceColor
				ui.StopHovering(id)
			}
			if police.MouseComponent.Clicked {
				police.Color = ui.PoliceColorSelected
				d.active = id
				d.wpEntity.Graphic.Hidden = false
				ui.StopHovering(id)
				return
			}
		}
	}

	// If we've selected a police unit, we can issue commands
	if d.active > 0 {
		police := police[d.active]

		if !d.submenuActive {

			// We can issue commands anywhere we want, as long as it's connected to roads.
			mX, mY := d.mouseTracker.MouseX, d.mouseTracker.MouseY
			mP := engo.Point{mX, mY}
			// Check which city is closest, and try to snap to that road
			nearest := CurrentMap.NearestNode(mP)
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
			var secondNearest *RouteNode
			for _, connected := range nearest.ConnectedTo {
				conn := CurrentMap.Node(connected)
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
				X: nearest.Location.X - waypointdx - ui.WaypointSize/2,
				Y: nearest.Location.Y - waypointdy - ui.WaypointSize/2,
			}

			// If we've snapped, we should create some kind of "waypoint-icon" player can click
			d.wpEntity.Graphic.SpaceComponent.Position = waypoint

			// Player can click, and will open submenu
			if d.wpEntity.MouseComponent.Clicked {
				// Using raw location because it's a HUD
				d.submenuTarget = waypoint
				d.showSubmenu(engo.Point{engo.Input.Mouse.X, engo.Input.Mouse.Y})
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
			police.Color = ui.PoliceColor
			d.active = 0
			ui.StopHovering(police.ID())
			d.wpEntity.Graphic.Hidden = true
			if d.submenuActive {
				d.hideSubmenu()
			}
		}
	}

	// Process all commands given to any units
	for _, p := range police {
		if p.CurrentCommand == CommandHold {
			p.CurrentCommand, p.CurrentTarget = p.processCommand()
		}
		switch p.CurrentCommand {
		case CommandHold:
		// Do nothing
		case CommandMove:
			if len(p.CurrentRoute.Nodes) < 1 {
				p.CurrentRoute = SetRoute(*p.Location, p.CurrentTarget, func(curr, goal, pos *RouteNode) float32 {
					dx := pos.Location.X - goal.Location.X
					dy := pos.Location.Y - goal.Location.Y
					dx2 := pos.Location.X - curr.Location.X
					dy2 := pos.Location.Y - curr.Location.Y
					return dx*dx + dy*dy + (dx2*dx2 + dy2*dy2)
				})
			}
			p.Move(dt)
		case CommandLookout:
			// If there's more to do, stop doing this and go do that other thing
			if len(p.Commands) > 0 {
				p.CurrentCommand = CommandHold
			}
			d.Lookout(p.PoliceComponent, p.CurrentTarget)
		case CommandSearchArea:
			// If there's more to do, stop doing this and go do that other thing
			if len(p.Commands) > 0 {
				p.CurrentCommand = CommandHold
			}
			p.Wander(dt, p.CurrentTarget)
			d.Lookout(p.PoliceComponent, p.CurrentTarget)
		case CommandTrafficControl:
			// If there's more to do, stop doing this and go do that other thing
			if len(p.Commands) > 0 {
				p.CurrentCommand = CommandHold
			}
		default:
			log.Println("Dunno what to do", p.CurrentCommand)
		}

		if p.CurrentResolve.BasicEntity != nil {
			engo.Mailbox.Dispatch(IncidentResolveMessage{p.CurrentResolve.IncidentComponent, p.CurrentResolve.BasicEntity})
			p.CurrentResolve = DispatchSystemIncidentEntity{}
		}
	}
}

func (d *DispatchSystem) Lookout(p *PoliceComponent, t engo.Point) {
	maxDist := p.Unit.ViewDistance

	if p.CurrentResolve.BasicEntity == nil {
		// Find new target, if any
		for _, incident := range incidents {
			if incident.Location.PointDistance(*p.Location) < maxDist {
				p.CurrentResolve = incident
				log.Println("Set currentResolve")
				break
			}
		}
	}
}
