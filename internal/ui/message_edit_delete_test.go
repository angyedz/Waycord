package ui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"

	"github.com/angyedz/Waycord/internal/core/attachments"
	"github.com/angyedz/Waycord/internal/ui/mock"
)

func TestEditOwnMessageAndRestoreComposerDraft(t *testing.T) {
	app := New()
	target := 4
	original := app.messages[target].Body
	originalTime, originalSentAt := app.messages[target].Time, app.messages[target].SentAt
	app.composer.SetText("draft before editing")
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	if !app.beginEdit(gtx, target) {
		t.Fatal("own text message was not editable")
	}
	if app.composer.Text() != original || app.editTarget != target {
		t.Fatalf("edit mode did not load message text: target=%d text=%q", app.editTarget, app.composer.Text())
	}
	app.composer.SetText("Updated locally")
	count := len(app.messages)
	app.sendPreview()
	if len(app.messages) != count || app.messages[target].Body != "Updated locally" || !app.messages[target].Edited || app.editTarget != -1 || app.messages[target].Time != originalTime || app.messages[target].SentAt != originalSentAt {
		t.Fatalf("edit created a new message or failed to update the target: %+v", app.messages[target])
	}
	if app.composer.Text() != "draft before editing" {
		t.Fatalf("composer draft was not restored after saving the edit: %q", app.composer.Text())
	}

	app.composer.SetText("keep this draft")
	if !app.beginEdit(gtx, target) {
		t.Fatal("edited message should remain editable")
	}
	app.cancelEdit()
	if app.composer.Text() != "keep this draft" || app.editTarget != -1 {
		t.Fatalf("cancel did not restore the draft: target=%d text=%q", app.editTarget, app.composer.Text())
	}
}

func TestEditRejectsOtherAuthorsAndEmptyChanges(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	if app.beginEdit(gtx, 2) {
		t.Fatal("message by another member was editable")
	}
	if !app.beginEdit(gtx, 4) {
		t.Fatal("own message was not editable")
	}
	app.composer.SetText(" \n ")
	app.sendPreview()
	if app.editTarget != 4 || app.messages[4].Body == "" || app.messages[4].Edited {
		// Empty edits remain in edit mode for correction rather than deleting a message.
		t.Fatalf("empty edit should remain open and keep the original body: target=%d body=%q", app.editTarget, app.messages[4].Body)
	}
}

func TestOnlyOwnTextMessagesExposeEditAction(t *testing.T) {
	for name, message := range map[string]mock.Message{
		"other author":   {Author: "mira", Body: "hello"},
		"own text":       {Author: "You", Self: true, Body: "hello"},
		"own attachment": {Author: "You", Self: true, Attachment: &attachments.Attachment{Name: "photo.png"}},
		"own media":      {Author: "You", Self: true, Media: &mock.Media{Name: "wave", Emoji: "👋", Kind: mock.MediaGIF}},
	} {
		want := name == "own text"
		if got := canEditMessage(message); got != want {
			t.Errorf("canEditMessage(%s) = %t, want %t", name, got, want)
		}
	}
}

func TestOwnMessageContextMenuStartsEditAndDeleteFlows(t *testing.T) {
	app := New()
	const target = 4
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(190, 240)}}
	app.contextTarget = target
	app.contextMenu(gtx)
	app.contextEdit.Click()
	app.processContextMenu(gtx)
	if app.editTarget != target || app.contextTarget != -1 {
		t.Fatal("edit menu item did not enter edit mode and close the menu")
	}
	app.cancelEdit()
	app.contextTarget = target
	app.contextMenu(gtx)
	app.contextDelete.Click()
	app.processContextMenu(gtx)
	if !app.deleteDialogOpen || app.deleteTarget != target || len(app.messages) != len(mock.Messages) {
		t.Fatal("delete menu item should open confirmation without removing the message")
	}
}

func TestContextMenuKeepsItsHeightDuringCloseAnimation(t *testing.T) {
	app := New()
	app.contextTarget = 2 // another member's message
	if got := app.contextMenuHeight(); got != 144 {
		t.Fatalf("other member menu height = %d, want 144", got)
	}
	app.contextTarget, app.contextShown = 4, 4
	if got := app.contextMenuHeight(); got != 216 {
		t.Fatalf("own text menu height = %d, want 216", got)
	}
	app.contextTarget = -1
	if got := app.contextMenuHeight(); got != 216 {
		t.Fatalf("closing own menu jumped to %d, want stable height 216", got)
	}
}

func TestDeleteRequiresConfirmationAndShiftsRepliesAndReactions(t *testing.T) {
	app := New()
	app.messages = []mock.Message{
		{Author: "mira", Body: "source"},
		{Author: "You", Self: true, Body: "delete me"},
		{Author: "You", Self: true, Body: "reply to deleted", HasReply: true, ReplyTo: 1},
		{Author: "You", Self: true, Body: "source two"},
		{Author: "You", Self: true, Body: "reply to source two", HasReply: true, ReplyTo: 3},
	}
	app.reactions.Toggle(0, "❤️")
	app.reactions.Toggle(1, "🔥")
	app.reactions.Toggle(2, "✅")
	app.reactions.Toggle(3, "😄")
	app.reactions.Toggle(4, "🎉")
	app.reactionWidgets[2] = &reactionWidget{}
	app.reactionWidgets[4] = &reactionWidget{}
	if app.requestDelete(0) {
		t.Fatal("request delete accepted a non-self message")
	}
	if !app.requestDelete(1) || !app.deleteDialogOpen || len(app.messages) != 5 {
		t.Fatal("requesting delete should open confirmation without removing the message")
	}
	app.cancelDelete()
	if app.deleteDialogOpen || len(app.messages) != 5 {
		t.Fatal("cancel should close confirmation and keep the message")
	}
	app.replyTarget = 1
	app.requestDelete(1)
	app.confirmDelete()
	if app.deleteDialogOpen || len(app.messages) != 4 || app.replyTarget != -1 || app.messages[1].HasReply || !app.messages[3].HasReply || app.messages[3].ReplyTo != 2 {
		t.Fatalf("delete did not remove the target and repair reply references: %+v", app.messages)
	}
	if !app.reactions.Has(0, "❤️") || !app.reactions.Has(1, "✅") || !app.reactions.Has(2, "😄") || !app.reactions.Has(3, "🎉") || app.reactions.Has(1, "🔥") {
		t.Fatal("reactions were lost or shifted to the wrong messages")
	}
	if len(app.reactionWidgets) != 0 {
		t.Fatal("message-specific hover widgets were not invalidated after deletion")
	}
}

func TestDeleteConfirmationButtonsWork(t *testing.T) {
	app := New()
	if !app.requestDelete(4) {
		t.Fatal("could not open delete confirmation for own message")
	}
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(480, 300)}}
	app.deleteConfirmation(gtx, gtx.Constraints.Max)
	app.deleteCancel.Click()
	if !app.processDeleteConfirmation(gtx) || app.deleteDialogOpen || len(app.messages) != len(mock.Messages) {
		t.Fatal("cancel button did not safely dismiss delete confirmation")
	}
	if !app.requestDelete(4) {
		t.Fatal("could not reopen delete confirmation")
	}
	app.deleteConfirmation(gtx, gtx.Constraints.Max)
	app.deleteConfirm.Click()
	if !app.processDeleteConfirmation(gtx) || app.deleteDialogOpen || len(app.messages) != len(mock.Messages)-1 {
		t.Fatal("confirm button did not delete the selected message")
	}
}
