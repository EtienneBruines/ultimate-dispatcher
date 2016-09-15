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
	Incident *Incident
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

type IncidentSpawningSystem struct {
	world         *ecs.World
	incidentLabel ui.Label

	activeIncidents []*Incident
}

func (d *IncidentSpawningSystem) New(w *ecs.World) {
	d.world = w

	engo.Mailbox.Listen("IncidentDebugViewMessage", func(m engo.Message) {
		debugMsg := m.(IncidentDebugViewMessage)

		if debugMsg.NewValue {
			// TODO: Everything should become visible now!
		} else {
			// TODO: Everything should become invisible now!
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

func (d *IncidentSpawningSystem) Remove(b ecs.BasicEntity) {}

func (d *IncidentSpawningSystem) Update(dt float32) {
	d.incidentLabel.SetText(fmt.Sprintf("Active Incidents: %d", len(d.activeIncidents)))
}

func (d *IncidentSpawningSystem) Spawn(in *Incident) {
	loc := CurrentMap.NearestNode(in.Location)

	type IncidentEntity struct {
		ecs.BasicEntity
		common.RenderComponent
		common.SpaceComponent
		common.MouseComponent
		IncidentComponent
	}

	ie := IncidentEntity{
		BasicEntity:     ecs.NewBasic(),
		RenderComponent: common.RenderComponent{Drawable: ui.IncidentGraphic, Color: ui.IncidentColor},
		SpaceComponent: common.SpaceComponent{
			Position: loc.Location,
			Width:    ui.IncidentSize,
			Height:   ui.IncidentSize,
			Rotation: 0,
		},
		IncidentComponent: IncidentComponent{*in},
	}
	ie.RenderComponent.SetZIndex(5)

	for _, system := range d.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent)
		case *common.MouseSystem:
			sys.Add(&ie.BasicEntity, &ie.MouseComponent, &ie.SpaceComponent, &ie.RenderComponent)
		case *IncidentDetailSystem:
			sys.Add(&ie.BasicEntity, &ie.RenderComponent, &ie.MouseComponent, &ie.IncidentComponent)
		case *DispatchSystem:
			sys.AddIncident(&ie.BasicEntity, &ie.RenderComponent, &ie.SpaceComponent, &ie.MouseComponent, &ie.IncidentComponent)
		}
	}

	d.activeIncidents = append(d.activeIncidents, in)
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
