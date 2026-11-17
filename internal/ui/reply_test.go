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

func TestHoverToolbarReplyStartsReply(t *testing.T) {
	app := New()
	const index = 2
	app.reactionWidget(index).hovered = true
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(600, 140)}}
	app.message(gtx, index, app.messages[index], false)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(550, 16)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(550, 16)})
	app.processReactions(gtx)
	if app.replyTarget != index {
		t.Fatalf("hover toolbar did not start a reply: %d", app.replyTarget)
	}
}
