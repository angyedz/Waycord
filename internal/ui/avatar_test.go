package ui

import (
	"errors"
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/angyedz/Waycord/internal/core/attachments"
)

func TestProfileAvatarButtonLoadsAndPersistsImage(t *testing.T) {
	app := New()
	initial := image.NewRGBA(image.Rect(0, 0, 2, 2))
	selected := image.NewRGBA(image.Rect(0, 0, 3, 3))
	var saved bool
	app.ConfigureLocalAvatar(initial, func() (*attachments.Attachment, error) {
		return &attachments.Attachment{Name: "avatar.png", Data: []byte("validated image data"), Preview: selected}, nil
	}, func(data []byte) error {
		if string(data) != "validated image data" {
			t.Fatalf("unexpected avatar bytes passed to storage: %q", data)
		}
		saved = true
		return nil
	})
	app.selfProfileOpen, app.selfProfileEditing = true, true
	app.selfAvatarButton.Click()
	app.processSelfProfile(layout.Context{})
	if !app.selfAvatarBusy {
		t.Fatal("clicking Change photo did not start the picker")
	}
	select {
	case result := <-app.selfAvatarResults:
		app.applySelfAvatarResult(result)
	case <-time.After(time.Second):
		t.Fatal("avatar file picker did not return")
	}
	if !saved || app.selfAvatar != selected || app.selfAvatarBusy || app.selfAvatarError != "" {
		t.Fatalf("avatar was not saved and applied: saved=%t image=%p busy=%t error=%q", saved, app.selfAvatar, app.selfAvatarBusy, app.selfAvatarError)
	}
}

func TestProfileAvatarKeepsItsCompactSquareSize(t *testing.T) {
	app := New()
	app.ConfigureLocalAvatar(image.NewRGBA(image.Rect(0, 0, 8, 8)), nil, nil)
	var ops op.Ops
	gtx := layout.Context{
		Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(100, 80)},
	}
	dimensions := app.profileAvatar(40, 14)(gtx)
	if dimensions.Size != image.Pt(40, 40) {
		t.Fatalf("avatar size = %v, want 40x40", dimensions.Size)
	}
}

func TestProfileAvatarDoesNotReplaceCurrentImageWhenSaveFails(t *testing.T) {
	app := New()
	current := image.NewRGBA(image.Rect(0, 0, 2, 2))
	app.ConfigureLocalAvatar(current, nil, func([]byte) error { return errors.New("disk full") })
	app.applySelfAvatarResult(selfAvatarResult{avatar: &attachments.Attachment{
		Data: []byte("valid bytes"), Preview: image.NewRGBA(image.Rect(0, 0, 3, 3)),
	}})
	if app.selfAvatar != current || app.selfAvatarError == "" {
		t.Fatal("failed avatar persistence should preserve the current image and show an error")
	}
}

func TestProfileAvatarRejectsFilesWithoutAnImagePreview(t *testing.T) {
	app := New()
	current := image.NewRGBA(image.Rect(0, 0, 2, 2))
	app.ConfigureLocalAvatar(current, nil, func([]byte) error {
		t.Fatal("invalid avatar data reached persistent storage")
		return nil
	})
	app.applySelfAvatarResult(selfAvatarResult{avatar: &attachments.Attachment{Data: []byte("text file")}})
	if app.selfAvatar != current || app.selfAvatarError == "" {
		t.Fatal("non-image attachment should not replace the current avatar")
	}
}
