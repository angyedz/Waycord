package ui

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestRightClickTargetsMessage(t *testing.T) {
	app := New()
	const index = 2
	app.reactionWidget(index)
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Min: image.Pt(600, 0), Max: image.Pt(600, 300)}}
	app.processContextMenu(gtx)
	app.message(gtx, index, app.messages[index])
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Move, Position: f32.Pt(120, 10)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonSecondary, Position: f32.Pt(120, 10)})
	app.processContextMenu(gtx)
	if app.contextTarget != index {
		t.Fatalf("right-click selected %d, want %d", app.contextTarget, index)
	}
}
