package ui

import "testing"

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
