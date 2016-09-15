package ui

import (
	"engo.io/engo/common"
	"image/color"
)

const (
	NodeSize           float32 = 10
	RoadSize           float32 = 2
	IncidentSize       float32 = 2 * NodeSize
	IncidentReportSize float32 = 1 * NodeSize
	PoliceSize         float32 = 2 * NodeSize
	WaypointSize       float32 = 1 * NodeSize

	TooltipLineHeight float32 = 24
	PoliceZIndex      float32 = 500
)

var (
	NodeColor                = color.NRGBA{0, 255, 255, 255}
	RoadColor                = color.White
	IncidentColor            = color.NRGBA{0, 0, 0, 255}
	IncidentColorHover       = color.NRGBA{153, 0, 0, 255}
	IncidentReportColor      = color.NRGBA{255, 0, 0, 128}
	IncidentReportColorHover = color.NRGBA{153, 0, 0, 255}
	PoliceColor              = color.NRGBA{0, 0, 255, 180}
	PoliceColorSelected      = color.NRGBA{255, 0, 255, 255}
	PoliceColorHover         = color.NRGBA{0, 255, 255, 255}
	TooltipColor             = color.NRGBA{230, 230, 230, 240}
	TooltipColorHover        = color.NRGBA{230, 230, 180, 255}
	TooltipColorBorder       = color.Black
	WaypointColor            = color.NRGBA{0, 255, 0, 150}

	NodeGraphic           = common.Circle{}
	RoadGraphic           = common.Rectangle{}
	IncidentGraphic       = common.Circle{}
	IncidentReportGraphic = common.Circle{}
	PoliceGraphic         = common.Circle{}
	TooltipGraphic        = common.Rectangle{BorderWidth: 1, BorderColor: TooltipColorBorder}
	WaypointGraphic       = common.Rectangle{}
)
