// Package reactions stores local preview reactions by stable message index.
package reactions

type Store struct{ items map[int]map[string]bool }

func (s *Store) Has(message int, emoji string) bool { return s.items[message][emoji] }

// Toggle adds or removes the current user's reaction.
func (s *Store) Toggle(message int, emoji string) {
	if emoji == "" {
		return
	}
	if s.items == nil {
		s.items = make(map[int]map[string]bool)
	}
	if s.items[message] == nil {
		s.items[message] = make(map[string]bool)
	}
	if s.items[message][emoji] {
		delete(s.items[message], emoji)
	} else {
		s.items[message][emoji] = true
	}
	if len(s.items[message]) == 0 {
		delete(s.items, message)
	}
}
