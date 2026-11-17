package ui

import "testing"

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
