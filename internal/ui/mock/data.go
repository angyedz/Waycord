// Package mock is an explicit, offline-only fixture source for the design
// preview. It never authenticates to or sends requests to Discord.
package mock

import (
	"time"

	"github.com/angyedz/Waycord/internal/core/messages"
)

type Message = messages.Message
type Media = messages.Media
type MediaKind = messages.MediaKind

const (
	MediaGIF     = messages.MediaGIF
	MediaSticker = messages.MediaSticker
)

var Channels = []string{"general", "dev-chat", "showcase", "ideas"}

type Category struct {
	Name     string
	Channels []int
}

type Member struct {
	Name     string
	Role     string
	Status   string
	Activity string
	Bot      bool
}

type Event struct {
	Name     string
	When     string
	Location string
}

type Guide struct {
	Steps []string
}

type Role struct {
	Name    string
	Members int
}

type BoostInfo struct {
	Level      int
	Count      int
	NextLevel  int
	Supporters int
}

type Server struct {
	Name          string
	Icon          string
	Channels      []string
	Uncategorized []int
	Categories    []Category
	VoiceChannel  string
	Members       []Member
	Events        []Event
	Guide         *Guide
	Roles         []Role
	Boosts        *BoostInfo
}

var Servers = []Server{
	{
		Name: "Waycord", Icon: "W", Channels: Channels, Uncategorized: []int{0, 1, 2, 3}, VoiceChannel: "Lounge",
		Members: []Member{{Name: "mira", Role: "Members", Status: "Online", Activity: "Listening to music"}, {Name: "angyedz", Role: "Administrators", Status: "Online"}, {Name: "k66", Role: "Members", Status: "Idle"}, {Name: "GSL bot", Role: "Bots", Status: "Online", Bot: true}},
		Guide:   &Guide{Steps: []string{"Choose a channel to start chatting", "Say hello to the community", "Share something you are working on"}},
		Events:  []Event{{Name: "Community meetup", When: "Nov 21 · 18:00", Location: "Lounge"}, {Name: "Showcase night", When: "Nov 28 · 19:30", Location: "showcase"}},
		Roles:   []Role{{Name: "Administrators", Members: 1}, {Name: "Members", Members: 2}, {Name: "Bots", Members: 1}},
	},
	{
		Name: "Studio", Icon: "S", Channels: []string{"welcome", "chat", "projects", "feedback"}, Uncategorized: []int{0},
		Categories: []Category{{Name: "Community", Channels: []int{1}}, {Name: "Workshop", Channels: []int{2, 3}}},
		Members:    []Member{{Name: "mira", Role: "Moderators", Status: "Online"}, {Name: "k66", Role: "Members", Status: "Idle"}},
		Roles:      []Role{{Name: "Hosts", Members: 1}, {Name: "Members", Members: 1}},
		Boosts:     &BoostInfo{Level: 2, Count: 5, NextLevel: 7, Supporters: 3},
	},
}

var Messages = []Message{
	{Direct: true, Author: "mira", Time: "10:24", Body: "Hey! Are we still meeting later?"},
	{Direct: true, Author: "You", Time: "10:26", Body: "Yes, see you there!", Self: true},
	{Channel: 1, Author: "mira", Time: "10:24", SentAt: time.Date(2026, 11, 17, 10, 24, 8, 0, time.UTC), Body: "Hey everyone! How is your day going?", Pinned: true},
	{Channel: 1, Author: "mira", Time: "10:24", SentAt: time.Date(2026, 11, 17, 10, 24, 22, 0, time.UTC), Body: "I've got a few ideas to share."},
	{Channel: 1, Author: "angyedz", Time: "10:26", Body: "Pretty good! I have **bold**, *italic*, and `code` styles.", Self: true},
	{Channel: 1, Author: "k66", Time: "10:27", Body: "I have been putting together a few ideas."},
	{Channel: 1, Author: "mira", Time: "10:31", Body: "Nice, share them when you get a chance!"},
	{Channel: 1, Author: "angyedz", Time: "10:35", Body: "Will do! 😊", Self: true},
	{Channel: 0, Author: "mira", Time: "09:00", Body: "Welcome, everyone!"},
	{Channel: 0, Author: "angyedz", Time: "09:04", Body: "Glad to be here!", Self: true},
	{Channel: 0, Author: "k66", Time: "09:12", Body: "Good morning! What are you all up to today?"},
	{Channel: 0, Author: "mira", Time: "09:18", Body: "Taking a little break and listening to music ☕"},
	{Channel: 0, Author: "angyedz", Time: "09:21", Body: "That sounds perfect 😊", Self: true},
	{Channel: 0, Author: "k66", Time: "09:25", Body: "Any recommendations?"},
	{Channel: 2, Author: "k66", Time: "Yesterday", Body: "Here is something I made this week."},
	{Channel: 3, Author: "mira", Time: "Yesterday", Body: "What should we try next?"},
	{Server: 1, Channel: 0, Author: "mira", Time: "10:00", Body: "Welcome to the studio!"},
	{Server: 1, Channel: 2, Author: "k66", Time: "10:15", Body: "Share your work here."},
}
