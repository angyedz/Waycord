package ui

import (
	"image"
	"os"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/ota"
	"github.com/angyedz/Waycord/internal/ui/mock"
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

func TestMessagesGroupOnlySameAuthorConversationAndMinute(t *testing.T) {
	stamp := time.Date(2026, 11, 17, 10, 24, 5, 0, time.UTC)
	a := mock.Message{Author: "mira", Channel: 1, SentAt: stamp}
	b := a
	b.SentAt = stamp.Add(40 * time.Second)
	if !messagesGroup(a, b) {
		t.Fatal("messages from the same minute should group")
	}
	for name, change := range map[string]func(*mock.Message){
		"author":       func(m *mock.Message) { m.Author = "other" },
		"channel":      func(m *mock.Message) { m.Channel++ },
		"direct":       func(m *mock.Message) { m.Direct = true },
		"minute":       func(m *mock.Message) { m.SentAt = stamp.Add(time.Minute) },
		"day":          func(m *mock.Message) { m.SentAt = stamp.Add(24 * time.Hour) },
		"unknown time": func(m *mock.Message) { m.SentAt = time.Time{} },
	} {
		candidate := b
		change(&candidate)
		if messagesGroup(a, candidate) {
			t.Errorf("messages grouped across %s", name)
		}
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

func TestFormattedMessageRemainsSelectableWithoutMarkdownMarkers(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(700, 400)}}
	message := mock.Message{Body: "**bold** and *italic*"}
	app.messageContent(gtx, 0, message)
	body := &app.reactionWidget(0).body
	if got := body.Text(); got != "bold and italic" {
		t.Fatalf("selectable text contains markdown syntax: %q", got)
	}
	body.SetCaret(0, 4)
	if got := body.SelectedText(); got != "bold" {
		t.Fatalf("unexpected formatted selection: %q", got)
	}
	ops.Reset()
	app.messageContent(gtx, 0, message)
	if got := body.SelectedText(); got != "bold" {
		t.Fatalf("formatted selection was lost on relayout: %q", got)
	}
}

func TestUpdateCheckFailsClosedWithoutTrustedKey(t *testing.T) {
	app := New()
	app.checkForUpdates()
	if app.updateChecking || app.updateStatus != "Signed updates aren't enabled for this build yet." {
		t.Fatalf("update check did not fail closed: checking=%t status=%q", app.updateChecking, app.updateStatus)
	}
}

func TestUpdatePreferenceChangePersistsBeforeApplying(t *testing.T) {
	app := New()
	app.ConfigureUpdatePreferences("stable", false, func(string, bool) error { return os.ErrPermission })
	app.setUpdatePreferences("beta", true)
	if app.updateChannel != "stable" || app.updateAutoCheck {
		t.Fatalf("preferences changed after failed save: channel=%q auto=%t", app.updateChannel, app.updateAutoCheck)
	}
	if app.updateStatus == "" {
		t.Fatal("save failure was not shown")
	}
}

func TestUpdateResultFromPreviousChannelIsIgnored(t *testing.T) {
	app := New()
	app.updateChannel = "beta"
	app.updateChecking = true
	app.updateResult <- updateCheckResult{channel: "stable", update: ota.Update{Available: true}}
	app.processUpdateChecks()
	if app.updateChecking || app.availableUpdate != nil || app.updateStatus != "" {
		t.Fatalf("old channel result changed the update screen: checking=%t update=%+v status=%q", app.updateChecking, app.availableUpdate, app.updateStatus)
	}
}
