// Package mock is an explicit, offline-only fixture source for the design
// preview. It never authenticates to or sends requests to Discord.
package mock

type Message struct {
	Channel int
	Author  string
	Time    string
	Body    string
	Self    bool
}

var Channels = []string{"general", "dev-chat", "showcase", "ideas"}
var Servers = []string{"W", "G", "O", "+"}
var Messages = []Message{
	{Channel: 1, Author: "mira", Time: "10:24", Body: "Waycord is looking clean already. Love the new sidebar."},
	{Channel: 1, Author: "angyedz", Time: "10:26", Body: "Go + Gio. No browser process, no Electron.", Self: true},
	{Channel: 1, Author: "k66", Time: "10:27", Body: "The update system needs a rollback. No exceptions."},
	{Channel: 1, Author: "mira", Time: "10:31", Body: "Signed releases + a lightweight launcher could do it."},
	{Channel: 1, Author: "angyedz", Time: "10:35", Body: "Already working on that. 🔥", Self: true},
	{Channel: 0, Author: "mira", Time: "09:00", Body: "Welcome to the Waycord offline preview!"},
	{Channel: 0, Author: "angyedz", Time: "09:04", Body: "This is demo data only. No Discord connection is active.", Self: true},
	{Channel: 2, Author: "k66", Time: "Yesterday", Body: "Dark mode, compact panels, quiet accents."},
	{Channel: 3, Author: "mira", Time: "Yesterday", Body: "Ideas: virtualized lists, low CPU idle, native voice."},
}
