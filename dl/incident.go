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

type Incident interface {
	Type() string
	Update(float32)
	SetLocation(*engo.Point)

	// Resolved indicates whether or not the incident has been resolved. < 0 if failed, > 0 if succeeded, 0 if ongoing.
	Resolved() int
	Reward() int
	Penalty() int
}

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

type IncidentResolveMessage struct {
	Incident *IncidentComponent
	Basic    *ecs.BasicEntity
}

func (IncidentResolveMessage) Type() string { return "IncidentResolveMessage" }

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

type UrgencyLevel uint8

const (
	UrgencyCritical UrgencyLevel = iota
	UrgencyUrgent
	UrgencyNeutral
	UrgencyNotUrgent
)

type IncidentComponent struct {
	Location *engo.Point
	Incident Incident

	Reports []IncidentReportComponent
}

type IncidentReportComponent struct {
	Location  *engo.Point
	Type string
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

type IncidentSystem struct {
	world         *ecs.World
	incidentLabel ui.Label

	activeIncidents       []*IncidentEntity
	activeIncidentReports map[uint64][]*IncidentReportEntity
}

func (d *IncidentSystem) New(w *ecs.World) {
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

	engo.Mailbox.Listen("IncidentResolveMessage", func(m engo.Message) {
		res := m.(IncidentResolveMessage)

		d.Resolve(res.Incident, res.Basic)
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

func (d *IncidentSystem) Remove(b ecs.BasicEntity) {
	for incidentID, reports := range d.activeIncidentReports {
		index := -1
		for i, report := range reports {
			if report.ID() == b.ID() {
				index = i
				break
			}
		}
		if index >= 0 {
			d.activeIncidentReports[incidentID] = append(reports[:index], reports[index+1:]...)
			return
		}
	}

	if reports, ok := d.activeIncidentReports[b.ID()]; ok {
		var ids []ecs.BasicEntity
		for _, report := range reports {
			ids = append(ids, report.BasicEntity)
		}
		for _, id := range ids {
			d.world.RemoveEntity(id)
		}
		delete(d.activeIncidentReports, b.ID())
	}

	index := -1
	for i, incident := range d.activeIncidents {
		if incident.ID() == b.ID() {
			index = i
			break
		}
	}
	if index >= 0 {
		d.activeIncidents = append(d.activeIncidents[:index], d.activeIncidents[index+1:]...)
	}
}

func (d *IncidentSystem) Update(dt float32) {
	d.incidentLabel.SetText(fmt.Sprintf("Active Incidents: %d", len(d.activeIncidents)))

	// Manage all incidents
	var msgs []IncidentResolveMessage
	for _, i := range d.activeIncidents {
		i.Incident.Update(dt)

		f := i.Incident.Resolved()
		switch  {
		case f < 0:
			log.Println("You have failed, penalty", i.Incident.Penalty())
		case f > 0:
			log.Println("Good job! You gained", i.Incident.Reward())
		default:
			continue
		}

		msgs = append(msgs, IncidentResolveMessage{&i.IncidentComponent, &i.BasicEntity})
	}

	// And remove any that can be removed
	for _, msg := range msgs {
		engo.Mailbox.Dispatch(msg)
	}
}

func (d *IncidentSystem) Spawn(in IncidentComponent) {
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
	in.Incident.SetLocation(&ie.SpaceComponent.Position)
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

func (d *IncidentSystem) Resolve(in *IncidentComponent, basic *ecs.BasicEntity) {
	log.Println("Resolving", in.Location)
	d.world.RemoveEntity(*basic)
	// TODO: "award" or "penalty"
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
