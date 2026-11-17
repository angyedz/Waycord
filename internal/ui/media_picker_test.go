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

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func TestSendMediaPreviewAddsGIFToCurrentConversation(t *testing.T) {
	app := New()
	app.activeServer, app.activeChannel = 1, 2
	app.emojiOpen = true
	media := mock.Media{Kind: mock.MediaGIF, Name: "Dancing cat", Emoji: "🐈"}

	app.sendMediaPreview(media)

	message := app.messages[len(app.messages)-1]
	if message.Server != 1 || message.Channel != 2 || message.Direct || message.Media == nil || *message.Media != media {
		t.Fatalf("media was not added to the current conversation: %+v", message)
	}
	if app.emojiOpen || !app.messageList.ScrollToEnd {
		t.Fatalf("sending media should close the picker and reveal the new message: open=%t scroll=%t", app.emojiOpen, app.messageList.ScrollToEnd)
	}
}

func TestSendMediaPreviewAddsStickerToDirectMessage(t *testing.T) {
	app := New()
	app.direct = true
	media := mock.Media{Kind: mock.MediaSticker, Name: "Happy blob", Emoji: "😊"}

	app.sendMediaPreview(media)

	message := app.messages[len(app.messages)-1]
	if !message.Direct || message.Media == nil || message.Media.Kind != mock.MediaSticker {
		t.Fatalf("sticker was not added to the direct conversation: %+v", message)
	}
}

func TestConversationSearchMatchesMediaName(t *testing.T) {
	app := New()
	app.sendMediaPreview(mock.Media{Kind: mock.MediaGIF, Name: "Dancing cat", Emoji: "🐈"})
	app.searchEditor.SetText("cat")

	visible := app.visibleMessages()
	if len(visible) != 1 || visible[0] != len(app.messages)-1 {
		t.Fatalf("search did not find the media label: %v", visible)
	}
}

func TestGIFComposerControlOpensGIFPickerTab(t *testing.T) {
	app := New()
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(100, 40)}}
	app.mediaTextControl(gtx, &app.gifButton, 1, "GIF", "Choose a GIF")
	router.Frame(&ops)
	router.Queue(
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(15, 14)},
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(15, 14)},
	)
	app.process(gtx)

	if !app.emojiOpen || app.mediaPickerTab != 1 {
		t.Fatalf("GIF control did not open the GIF picker: open=%t tab=%d", app.emojiOpen, app.mediaPickerTab)
	}
}

func TestMediaPickerTabsSwitchAndToggleClosed(t *testing.T) {
	app := New()
	app.openMediaPicker(1)
	if !app.emojiOpen || app.mediaPickerTab != 1 {
		t.Fatal("GIF tab did not open")
	}
	app.openMediaPicker(2)
	if !app.emojiOpen || app.mediaPickerTab != 2 {
		t.Fatal("switching picker controls did not open the selected tab")
	}
	app.openMediaPicker(2)
	if app.emojiOpen {
		t.Fatal("clicking the selected picker control did not close it")
	}
}

func TestGIFAndStickerMessagesHaveVisibleCards(t *testing.T) {
	for _, media := range []mock.Media{
		{Kind: mock.MediaGIF, Name: "Dancing cat", Emoji: "🐈"},
		{Kind: mock.MediaSticker, Name: "Happy blob", Emoji: "😊"},
	} {
		app := New()
		var ops op.Ops
		gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(480, 200)}}
		dimensions := app.mediaCard(gtx, &media)
		if dimensions.Size.X == 0 || dimensions.Size.Y == 0 {
			t.Fatalf("%v message rendered an empty media card: %+v", media.Kind, dimensions)
		}
	}
}

func TestMediaPickerTabsFitCompactPanel(t *testing.T) {
	app := New()
	for tab := range app.mediaTabButtons {
		app.mediaPickerTab = tab
		var ops op.Ops
		gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(480, 180)}}
		dimensions := app.emojiPicker(gtx)
		if dimensions.Size.X > 480 || dimensions.Size.Y > 180 || dimensions.Size.X == 0 || dimensions.Size.Y == 0 {
			t.Fatalf("media tab %d exceeded its layout bounds: %+v", tab, dimensions)
		}
	}
}
