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
var Messages = []Message{
	{Channel: 1, Author: "mira", Time: "10:24", Body: "Hey everyone! How is your day going?"},
	{Channel: 1, Author: "angyedz", Time: "10:26", Body: "Pretty good! What are you working on today?", Self: true},
	{Channel: 1, Author: "k66", Time: "10:27", Body: "I have been putting together a few ideas."},
	{Channel: 1, Author: "mira", Time: "10:31", Body: "Nice, share them when you get a chance!"},
	{Channel: 1, Author: "angyedz", Time: "10:35", Body: "Will do!", Self: true},
	{Channel: 0, Author: "mira", Time: "09:00", Body: "Welcome, everyone!"},
	{Channel: 0, Author: "angyedz", Time: "09:04", Body: "Glad to be here!", Self: true},
	{Channel: 2, Author: "k66", Time: "Yesterday", Body: "Here is something I made this week."},
	{Channel: 3, Author: "mira", Time: "Yesterday", Body: "What should we try next?"},
}
