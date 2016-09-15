package ui

import (
	"engo.io/engo/common"
	"image/color"
)

const (
	NodeSize     = 10
	RoadSize     = 2
	IncidentSize = 2 * NodeSize
	PoliceSize   = 2 * NodeSize
	WaypointSize = 1 * NodeSize

	TooltipLineHeight float32 = 24
	PoliceZIndex      float32 = 500
)

var (
	NodeColor           = color.NRGBA{0, 255, 255, 255}
	RoadColor           = color.White
	IncidentColor       = color.NRGBA{255, 0, 0, 128}
	IncidentColorHover  = color.NRGBA{153, 0, 0, 255}
	PoliceColor         = color.NRGBA{0, 0, 255, 128}
	PoliceColorSelected = color.NRGBA{255, 0, 255, 255}
	PoliceColorHover    = color.NRGBA{0, 0, 255, 255}
	TooltipColor        = color.NRGBA{230, 230, 230, 240}
	TooltipColorHover   = color.NRGBA{230, 230, 180, 255}
	TooltipColorBorder  = color.Black
	WaypointColor       = color.NRGBA{0, 255, 0, 150}

	NodeGraphic     = common.Circle{}
	RoadGraphic     = common.Rectangle{}
	IncidentGraphic = common.Circle{}
	PoliceGraphic   = common.Circle{}
	TooltipGraphic  = common.Rectangle{BorderWidth: 1, BorderColor: TooltipColorBorder}
	WaypointGraphic = common.Rectangle{}
)
