package dl

import (
	"fmt"
	"image/color"
	"log"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EtienneBruines/ultimate-dispatcher/ui"
)

var (
	// TODO: false for production
	debugView = true
)

const (
	incidentSpawningKey = "incident-spawning-key"
	incidentViewKey     = "incident-viewing-key"
)

type IncidentDebugViewMessage struct {
	NewValue bool
}

func (IncidentDebugViewMessage) Type() string { return "IncidentDebugViewMessage" }

type IncidentNewMessage struct {
	Incident IncidentComponent
}

func (IncidentNewMessage) Type() string { return "IncidentNewMessage" }

type IncidentDebugSystem struct {
	world *ecs.World
}

func (d *IncidentDebugSystem) New(w *ecs.World) {
	d.world = w
	engo.Input.RegisterButton(incidentSpawningKey, engo.F1)
	engo.Input.RegisterButton(incidentViewKey, engo.F2)
}

func (d *IncidentDebugSystem) Remove(b ecs.BasicEntity) {}

func (d *IncidentDebugSystem) Update(dt float32) {
	if engo.Input.Button(incidentSpawningKey).JustPressed() {
		// Spawn!
		log.Println("Spawn!")
	}

	if engo.Input.Button(incidentViewKey).JustPressed() {
		debugView = !debugView
		log.Println("IncidentDebugView:", debugView)
		engo.Mailbox.Dispatch(IncidentDebugViewMessage{debugView})
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

type IncidentComponent struct {
	Location *engo.Point
	Award    int
	Penalty  int
	Type     IncidentType

	Reports []IncidentReportComponent
}

type IncidentReportComponent struct {
	Location  *engo.Point
	Type      IncidentType
	MinAmount uint8
	MaxAmount uint8
	Urgency   UrgencyLevel
}

type IncidentEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	IncidentComponent
}

type IncidentReportEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	IncidentReportComponent
}

type IncidentSpawningSystem struct {
	world         *ecs.World
	incidentLabel ui.Label

	activeIncidents       []*IncidentEntity
	activeIncidentReports map[uint64][]*IncidentReportEntity
}

func (d *IncidentSpawningSystem) New(w *ecs.World) {
	d.world = w
	d.activeIncidentReports = make(map[uint64][]*IncidentReportEntity)

	engo.Mailbox.Listen("IncidentDebugViewMessage", func(m engo.Message) {
		debugMsg := m.(IncidentDebugViewMessage)

		for _, incident := range d.activeIncidents {
			incident.RenderComponent.Hidden = !debugMsg.NewValue
		}
	})

	engo.Mailbox.Listen("IncidentNewMessage", func(m engo.Message) {
		newMsg := m.(IncidentNewMessage)

		d.Spawn(newMsg.Incident)
	})

	// Show the incident counter in the corner
	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.White,
		Size: 48,
	}
	if err := fnt.CreatePreloaded(); err != nil {
		panic(err)
	}

	d.incidentLabel = ui.Label{
		BasicEntity:     ecs.NewBasic(),
		Font:            fnt,
		SpaceComponent:  common.SpaceComponent{engo.Point{4, 4}, 100, 20, 0},
		RenderComponent: common.RenderComponent{Scale: engo.Point{0.5, 0.5}},
	}
	d.incidentLabel.SetText(fmt.Sprintf("Active Incidents: %d", len(d.activeIncidents)))
	d.incidentLabel.RenderComponent.SetShader(common.HUDShader)

	for _, system := range d.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&d.incidentLabel.BasicEntity, &d.incidentLabel.RenderComponent, &d.incidentLabel.SpaceComponent)
		}
	}
}

func (d *IncidentSpawningSystem) Remove(b ecs.BasicEntity) {
	delete(d.activeIncidentReports, b.ID())
	// TODO: also remove the Incident itself
	// TODO: also remove all incident reports!
}

func (d *IncidentSpawningSystem) Update(dt float32) {
	d.incidentLabel.SetText(fmt.Sprintf("Active Incidents: %d", len(d.activeIncidents)))
}

func (d *IncidentSpawningSystem) Spawn(in IncidentComponent) {
	ie := &IncidentEntity{
		BasicEntity: ecs.NewBasic(),
		RenderComponent: common.RenderComponent{
			Drawable:         ui.IncidentGraphic,
			Color:            ui.IncidentColor,
			Hidden:           !debugView,
			TextureAlignment: common.AlignCenter,
		},
		SpaceComponent: common.SpaceComponent{
			Position: *in.Location,
			Width:    ui.IncidentSize,
			Height:   ui.IncidentSize,
			Rotation: 0,
		},
	}
	in.Location = &ie.SpaceComponent.Position
	ie.IncidentComponent = in

	ie.RenderComponent.SetZIndex(5)

	for _, report := range in.Reports {
		re := &IncidentReportEntity{
			BasicEntity: ecs.NewBasic(),
			RenderComponent: common.RenderComponent{
				Drawable:         ui.IncidentReportGraphic,
				Color:            ui.IncidentReportColor,
				TextureAlignment: common.AlignCenter,
			},
			SpaceComponent: common.SpaceComponent{
				Position: *report.Location,
				Width:    ui.IncidentReportSize,
				Height:   ui.IncidentReportSize,
				Rotation: 0,
			},
			IncidentReportComponent: report,
		}
		report.Location = &re.SpaceComponent.Position
		re.RenderComponent.SetZIndex(5)

		var curList []*IncidentReportEntity
		if l, ok := d.activeIncidentReports[ie.ID()]; ok {
			curList = l
		}
		curList = append(curList, re)
		d.activeIncidentReports[ie.ID()] = curList
	}

	for _, system := range d.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent)
			for _, reports := range d.activeIncidentReports {
				for _, report := range reports {
					sys.Add(&report.BasicEntity, &report.RenderComponent, &report.SpaceComponent)
				}
			}
		case *common.MouseSystem:
			sys.Add(&ie.BasicEntity, &ie.MouseComponent, &ie.SpaceComponent, &ie.RenderComponent)
			for _, reports := range d.activeIncidentReports {
				for _, report := range reports {
					sys.Add(&report.BasicEntity, &report.MouseComponent, &report.SpaceComponent, &report.RenderComponent)
				}
			}
		case *IncidentDetailSystem:
			sys.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.MouseComponent, &ie.IncidentComponent)
		case *DispatchSystem:
			sys.AddIncident(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent, &ie.MouseComponent, &ie.IncidentComponent)
		}
	}

	d.activeIncidents = append(d.activeIncidents, ie)
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
