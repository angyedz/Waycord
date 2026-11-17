// Package messages owns local message rules shared by the UI and future
// protocol adapters. It has no dependency on Gio or Discord.
package messages

import (
	"strings"
	"time"

	"github.com/angyedz/Waycord/internal/core/attachments"
)

type MediaKind uint8

const (
	MediaGIF MediaKind = iota + 1
	MediaSticker
)

type Media struct {
	Kind  MediaKind
	Name  string
	Emoji string
}

type Message struct {
	Server     int
	Channel    int
	Author     string
	Time       string
	Body       string
	Self       bool
	Direct     bool
	HasReply   bool
	ReplyTo    int
	Pinned     bool
	Edited     bool
	SentAt     time.Time
	Attachment *attachments.Attachment
	Media      *Media
}

func CanEdit(message Message) bool {
	return message.Self && strings.TrimSpace(message.Body) != "" && message.Attachment == nil && message.Media == nil
}

func Edit(items []Message, index int, body string) bool {
	if index < 0 || index >= len(items) || !CanEdit(items[index]) {
		return false
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return false
	}
	items[index].Body = body
	items[index].Edited = true
	return true
}

func Delete(items *[]Message, index int) bool {
	if items == nil || index < 0 || index >= len(*items) || !(*items)[index].Self {
		return false
	}
	remaining := *items
	last := len(remaining) - 1
	copy(remaining[index:], remaining[index+1:])
	remaining[last] = Message{}
	remaining = remaining[:last]
	for message := range remaining {
		if !remaining[message].HasReply {
			continue
		}
		switch {
		case remaining[message].ReplyTo == index:
			remaining[message].HasReply = false
			remaining[message].ReplyTo = 0
		case remaining[message].ReplyTo > index:
			remaining[message].ReplyTo--
		}
	}
	*items = remaining
	return true
}
