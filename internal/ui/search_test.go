package ui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestSearchEditorIsSingleLine(t *testing.T) {
	app := New()
	if !app.searchEditor.SingleLine {
		t.Fatal("channel search should use a single-line editor")
	}
}

func TestCenteredSearchInputCentersItsBaseline(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(320, 42)},
	}
	dims := centeredSearchField(gtx, func(gtx layout.Context) layout.Dimensions {
		return fixedHeight(36, func(gtx layout.Context) layout.Dimensions {
			return centeredSearchInput(gtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(180, 18), Baseline: 13}
			})
		})(gtx)
	})
	if dims.Size.Y != 42 || dims.Baseline != 25 {
		t.Fatalf("search line should be centered through the 42px header field, got size=%v baseline=%d", dims.Size, dims.Baseline)
	}
}

func TestSearchFiltersCurrentConversationAndPinnedMessages(t *testing.T) {
	app := New()
	app.searchEditor.SetText("ideas")
	visible := app.visibleMessages()
	if len(visible) != 2 || visible[0] != 3 || visible[1] != 5 {
		t.Fatalf("search returned the wrong messages: %v", visible)
	}
	app.pinsOnly = true
	app.searchEditor.SetText("hey")
	visible = app.visibleMessages()
	if len(visible) != 1 || visible[0] != 2 {
		t.Fatalf("search did not intersect with pinned messages: %v", visible)
	}
	app.direct = true
	app.pinsOnly = false
	app.searchEditor.SetText("meeting")
	visible = app.visibleMessages()
	if len(visible) != 1 || visible[0] != 0 {
		t.Fatalf("direct-message search escaped the active conversation: %v", visible)
	}
}

func TestConversationSwitchClearsScopedSearch(t *testing.T) {
	app := New()
	app.searchEditor.SetText("ideas")
	app.pinsOnly = true
	app.switchConversation(false, 1, 2)
	if app.searchEditor.Text() != "" || app.pinsOnly {
		t.Fatal("conversation switch retained a stale search or pinned filter")
	}
}
