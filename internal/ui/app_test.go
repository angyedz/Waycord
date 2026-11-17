package ui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

func TestPreviewMessageRemainsInItsConversation(t *testing.T) {
	app := New()
	app.direct = true
	app.composer.SetText("hello")
	app.sendPreview()
	message := app.messages[len(app.messages)-1]
	if !message.Direct || message.Body != "hello" {
		t.Fatalf("direct message lost its context: %+v", message)
	}
	if app.composer.Text() != "" {
		t.Fatal("composer was not cleared")
	}
	app.direct = false
	app.activeChannel = 2
	app.composer.SetText("hi channel")
	app.sendPreview()
	message = app.messages[len(app.messages)-1]
	if message.Direct || message.Channel != 2 {
		t.Fatalf("channel message lost its context: %+v", message)
	}
}

func TestSidebarKeepsWidthWhenToggled(t *testing.T) {
	sidebar := newSidebarState()
	sidebar.width = clampSidebarWidth(260)
	sidebar.toggle()
	if sidebar.shown {
		t.Fatal("sidebar did not collapse")
	}
	sidebar.toggle()
	if !sidebar.shown || sidebar.width != 260 {
		t.Fatalf("sidebar state was not restored: %+v", sidebar)
	}
	if clampSidebarWidth(0) != minSidebarWidth || clampSidebarWidth(1000) != maxSidebarWidth {
		t.Fatal("sidebar width limits not enforced")
	}
}

func TestPillDoesNotInheritContainerMinimumSize(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Min: image.Pt(200, 80), Max: image.Pt(200, 80)},
	}
	var button widget.Clickable
	dims := app.pill(gtx, &button, "Profile", true)
	if dims.Size.X >= 200 || dims.Size.Y >= 80 {
		t.Fatalf("button inherited container size: %v", dims.Size)
	}
}

func TestReactionChipRemainsCompact(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Min: image.Pt(200, 80), Max: image.Pt(300, 100)}}
	var button widget.Clickable
	size := app.reactionChip(gtx, &button, "🔥").Size
	if size.X > 60 || size.Y > 32 {
		t.Fatalf("reaction is too large: %v", size)
	}
}

func TestMessageSelectionSurvivesRelayout(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(700, 400)}}
	index := 2
	app.messageContent(gtx, index, app.messages[index])
	body := &app.reactionWidget(index).body
	body.SetCaret(0, 3)
	selected := body.SelectedText()
	if selected == "" {
		t.Fatal("message text was not selectable")
	}
	ops.Reset()
	app.messageContent(gtx, index, app.messages[index])
	if body.SelectedText() != selected {
		t.Fatal("selection was lost on relayout")
	}
}
