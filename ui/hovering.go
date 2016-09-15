package ui

import "engo.io/engo"

var hoverings = make(map[uint64]bool)

func StartHovering(uid uint64) {
	if len(hoverings) == 0 {
		engo.SetCursor(engo.CursorHand)
	}
	hoverings[uid] = true
}

func StopHovering(uid uint64) {
	delete(hoverings, uid)
	if len(hoverings) == 0 {
		engo.SetCursor(engo.CursorNone)
	}
}
