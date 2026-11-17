package ui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestMessageHistoryOnlyPinsToBottomWhenItOverflows(t *testing.T) {
	tests := []struct {
		name          string
		content, view int
		wantBottomPin bool
	}{
		{name: "short conversation", content: 180, view: 520, wantBottomPin: false},
		{name: "fills the viewport", content: 520, view: 520, wantBottomPin: false},
		{name: "long conversation", content: 900, view: 520, wantBottomPin: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := messageHistoryOverflows(test.content, test.view); got != test.wantBottomPin {
				t.Fatalf("messageHistoryOverflows(%d, %d) = %t, want %t", test.content, test.view, got, test.wantBottomPin)
			}
		})
	}
}

func TestShortConversationDoesNotLeaveMessagesPinnedAtBottom(t *testing.T) {
	app := New()
	app.direct = true
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(800, 600)},
	}
	app.chat(gtx)
	if app.messageList.Position.Length == 0 {
		t.Fatal("chat did not measure its initial history")
	}

	app.composer.SetText("A short local reply")
	app.sendPreview()
	ops.Reset()
	app.chat(gtx)
	if app.messageList.ScrollToEnd {
		t.Fatal("a short conversation should start at the top instead of leaving a blank area above messages")
	}
	if app.messageList.Position.First != 0 || app.messageList.Position.Offset != 0 {
		t.Fatalf("short history is not top-aligned: first=%d offset=%d", app.messageList.Position.First, app.messageList.Position.Offset)
	}
}

func TestLongConversationStillScrollsToLatestMessage(t *testing.T) {
	app := New()
	app.direct = true
	for i := 0; i < 80; i++ {
		app.messages = append(app.messages, app.messages[0])
		app.messages[len(app.messages)-1].Body = "A longer line of local preview text that occupies space in the conversation."
	}
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(800, 360)},
	}
	app.chat(gtx)
	if app.messageList.Position.Length <= gtx.Constraints.Max.Y {
		t.Fatalf("long history should exceed its viewport, got length %d", app.messageList.Position.Length)
	}
	app.messageList.Position.BeforeEnd = true
	app.composer.SetText("Latest message")
	app.sendPreview()
	if app.messageList.Position.BeforeEnd {
		t.Fatal("sending a message should clear the old away-from-end position")
	}
	ops.Reset()
	app.chat(gtx)
	if !app.messageList.ScrollToEnd {
		t.Fatal("a long conversation should remain pinned to its newest message after sending")
	}
}
