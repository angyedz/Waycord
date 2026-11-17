package ui

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"image"
	"os"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/core/attachments"
	"github.com/angyedz/Waycord/internal/ota"
	"github.com/angyedz/Waycord/internal/ui/mock"
	"github.com/angyedz/Waycord/internal/version"
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
		"server":       func(m *mock.Message) { m.Server++ },
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

func TestSwitchConversationKeepsDraftsAndServerChannel(t *testing.T) {
	app := New()
	app.composer.SetText("first server draft")
	app.switchConversation(false, 1, 2)
	if app.activeServer != 1 || app.activeChannel != 2 || app.composer.Text() != "" {
		t.Fatal("server switch did not select the requested conversation")
	}
	app.composer.SetText("second server draft")
	app.switchConversation(true, 1, 2)
	app.composer.SetText("private draft")
	app.switchConversation(false, 0, 1)
	if app.composer.Text() != "first server draft" || app.serverChannels[1] != 2 {
		t.Fatal("server draft or selected channel was not restored")
	}
	app.switchConversation(false, 1, app.serverChannels[1])
	if app.composer.Text() != "second server draft" {
		t.Fatal("second server draft was not restored")
	}
	app.switchConversation(true, 1, 2)
	if app.composer.Text() != "private draft" {
		t.Fatal("direct conversation draft was not restored")
	}
}

func TestReplyIsAttachedToSentMessageAndClearedOnSwitch(t *testing.T) {
	app := New()
	app.replyTarget = 2
	app.composer.SetText("Thanks, I will take a look.")
	app.sendPreview()
	last := app.messages[len(app.messages)-1]
	if !last.HasReply || last.ReplyTo != 2 || app.replyTarget != -1 {
		t.Fatalf("reply was not sent with its target: %+v, pending=%d", last, app.replyTarget)
	}
	app.replyTarget = 2
	app.switchConversation(false, 1, 1)
	if app.replyTarget != -1 {
		t.Fatal("switching conversations left a stale reply target")
	}
}

func TestReplyPreviewAndQuotedMessageLayout(t *testing.T) {
	app := New()
	app.replyTarget = 2
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(600, 120)}}
	if dims := app.replyPreview(gtx); dims.Size.Y == 0 {
		t.Fatal("reply preview was not laid out")
	}
	quoted := mock.Message{HasReply: true, ReplyTo: 2, Author: "You", Body: "A reply"}
	ops.Reset()
	if dims := app.messageReply(gtx, len(app.messages), quoted); dims.Size.Y == 0 {
		t.Fatal("quoted message was not laid out")
	}
}

func TestMessageHoverDoesNotChangeRowHeight(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(600, 500)}}
	message := mock.Message{Author: "mira", Body: "Short message"}
	for _, grouped := range []bool{false, true} {
		app.reactionWidget(0).hovered = false
		ops.Reset()
		before := app.message(gtx, 0, message, grouped).Size
		app.reactionWidget(0).hovered = true
		ops.Reset()
		after := app.message(gtx, 0, message, grouped).Size
		if before != after {
			t.Fatalf("grouped=%t: hover changed dimensions from %v to %v", grouped, before, after)
		}
	}
}

func TestAttachmentDraftStaysInItsConversation(t *testing.T) {
	app := New()
	file := &attachments.Attachment{Name: "note.txt", Data: []byte("hello")}
	app.attachmentDrafts[app.conversationKey()] = file
	app.direct = true
	app.sendPreview()
	if got := len(app.messages); got != len(mock.Messages) {
		t.Fatalf("attachment from another conversation was sent: %d", got)
	}
	app.direct = false
	app.sendPreview()
	last := app.messages[len(app.messages)-1]
	if last.Attachment != file || last.Body != "" || app.attachmentDrafts[app.conversationKey()] != nil {
		t.Fatalf("attachment draft not sent and cleared: %+v", last)
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
	if app.updateChecking || app.updateStatus != "Updates are unavailable for this installation." {
		t.Fatalf("update check did not fail closed: checking=%t status=%q", app.updateChecking, app.updateStatus)
	}
}

func TestUpdateInstallProgressTransitionsToReadyState(t *testing.T) {
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	previousKey := version.SigningPublicKeyHex
	version.SigningPublicKeyHex = hex.EncodeToString(public)
	defer func() { version.SigningPublicKeyHex = previousKey }()

	app := New()
	app.availableUpdate = &ota.Update{Available: true, Payload: ota.Payload{Version: "v1.0.1", Size: 10}}
	app.updateInvalidate = func() {}
	finished := make(chan struct{})
	app.ConfigureUpdateInstallation(func(_ context.Context, _ ota.Update, channel string, key ed25519.PublicKey, progress ota.ProgressFunc) error {
		if channel != "stable" || string(key) != string(public) {
			t.Errorf("installer received wrong release selection or key: channel=%q", channel)
		}
		progress(4, 10)
		progress(10, 10)
		close(finished)
		return nil
	}, func() error { return nil })

	app.installAvailableUpdate()
	if !app.updateInstalling || app.updateTotal != 10 {
		t.Fatalf("download did not enter progress state: installing=%t total=%d", app.updateInstalling, app.updateTotal)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("installer was not called")
	}
	for deadline := time.Now().Add(time.Second); len(app.updateInstallResult) == 0 && time.Now().Before(deadline); {
		time.Sleep(time.Millisecond)
	}
	app.processUpdateChecks()
	if app.updateInstalling || !app.updateStaged || app.updateStatus != "Update ready to install. Restart Waycord when you're ready." {
		t.Fatalf("download did not transition to ready state: installing=%t staged=%t status=%q", app.updateInstalling, app.updateStaged, app.updateStatus)
	}
	if app.updateDownloaded != 10 || app.updateProgressLabel() != "Downloading update… 100%" {
		t.Fatalf("final progress wasn't rendered: downloaded=%d label=%q", app.updateDownloaded, app.updateProgressLabel())
	}
}

func TestUpdateInstallRequiresManagedRestartPath(t *testing.T) {
	app := New()
	app.ConfigureUpdateInstallation(func(context.Context, ota.Update, string, ed25519.PublicKey, ota.ProgressFunc) error { return nil }, nil)
	if app.updateInstaller != nil || app.restartUpdatedApp != nil {
		t.Fatal("update install was enabled without a managed restart path")
	}
}

func TestStagedUpdateRemainsVisibleAfterChannelChange(t *testing.T) {
	app := New()
	app.availableUpdate = &ota.Update{Available: true, Payload: ota.Payload{Version: "v1.0.1"}}
	app.updateStaged = true
	app.setUpdatePreferences("beta", false)
	if app.availableUpdate == nil || !app.updateStaged {
		t.Fatal("changing update channel hid a staged restart")
	}
}

func TestCategoryDisclosureIconsLayoutWithoutFontGlyphs(t *testing.T) {
	app := New()
	if app.categoryExpandedIcon == nil || app.categoryCollapsedIcon == nil {
		t.Fatal("category disclosure icons were not initialized")
	}
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(320, 600)}}
	dimensions := app.serverChannelsLayout(gtx)
	if dimensions.Size.X == 0 || dimensions.Size.Y == 0 {
		t.Fatalf("category sidebar did not layout: %+v", dimensions)
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
	app.updateChannel = "stable"
	app.updateCheckID = 2
	app.updateChecking = true
	app.updateResult <- updateCheckResult{id: 1, channel: "stable", update: ota.Update{Available: true}}
	app.processUpdateChecks()
	if !app.updateChecking || app.availableUpdate != nil || app.updateStatus != "" {
		t.Fatalf("old channel result changed the update screen: checking=%t update=%+v status=%q", app.updateChecking, app.availableUpdate, app.updateStatus)
	}
}
