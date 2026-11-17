// Package protocol defines the boundary between Waycord's application core
// and a future Discord transport. It contains contracts only; no network or
// authentication implementation is included in the offline preview.
package protocol

import (
	"context"
	"io"
	"time"
)

// Discord IDs stay strings so snowflakes are never rounded by numeric types.
type ServerID string
type ChannelID string
type MessageID string
type UserID string

// ConversationRef identifies a text channel. ServerID is empty for a DM.
type ConversationRef struct {
	ServerID  ServerID
	ChannelID ChannelID
}

type Author struct {
	ID       UserID
	Username string
}

type Attachment struct {
	ID          string
	Filename    string
	ContentType string
	Size        int64
	URL         string
}

type Reaction struct {
	Emoji string
	Count int
	Me    bool
}

type Message struct {
	ID          MessageID
	Author      Author
	Content     string
	CreatedAt   time.Time
	EditedAt    *time.Time
	ReplyTo     *MessageID
	Attachments []Attachment
	Reactions   []Reaction
}

// MessagePage is ordered oldest to newest. When HasMore is true, NextBefore
// identifies the oldest item in this page for the next history request.
type MessagePage struct {
	Items      []Message
	NextBefore MessageID
	HasMore    bool
}

type Upload struct {
	Filename    string
	ContentType string
	Size        int64
	Open        func() (io.ReadCloser, error)
}

type SendInput struct {
	Content     string
	ReplyTo     *MessageID
	Attachments []Upload
}

type EditInput struct {
	Content string
}

// MessageService is the smallest useful messaging contract for application
// services. Implementations return errors that may wrap the standard sentinel
// values below; transport details remain private to the adapter.
type MessageService interface {
	ListMessages(context.Context, ConversationRef, MessageID, int) (MessagePage, error)
	SendMessage(context.Context, ConversationRef, SendInput) (Message, error)
	EditMessage(context.Context, ConversationRef, MessageID, EditInput) (Message, error)
	DeleteMessage(context.Context, ConversationRef, MessageID) error
	SetReaction(context.Context, ConversationRef, MessageID, string, bool) error
}
