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
	app.message(gtx, index, app.messages[index], false)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Move, Position: f32.Pt(120, 10)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonSecondary, Position: f32.Pt(120, 10)})
	app.processContextMenu(gtx)
	if app.contextTarget != index {
		t.Fatalf("right-click selected %d, want %d", app.contextTarget, index)
	}
}

func TestContextMenuReplyStartsLocalReply(t *testing.T) {
	app := New()
	app.contextTarget = 2
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(190, 130)}}
	app.contextMenu(gtx)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(48, 20)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(48, 20)})
	app.processContextMenu(gtx)
	if app.replyTarget != 2 || app.contextTarget != -1 {
		t.Fatalf("reply action did not select and close the menu: target=%d menu=%d", app.replyTarget, app.contextTarget)
	}
}

func TestContextMenuPinActionAddsMessageToPinnedFilter(t *testing.T) {
	app := New()
	const target = 3
	app.contextTarget = target
	if app.messages[target].Pinned {
		t.Fatal("fixture message should start unpinned for this test")
	}
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(190, 150)}}
	app.contextMenu(gtx)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(55, 115)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(55, 115)})
	app.processContextMenu(gtx)
	app.pinsOnly = true
	visible := app.visibleMessages()
	if !app.messages[target].Pinned || len(visible) != 2 || visible[1] != target {
		t.Fatalf("pin action did not update the pinned view: pinned=%t visible=%v", app.messages[target].Pinned, visible)
	}
}
