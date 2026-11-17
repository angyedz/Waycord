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

// RemoveAndShift discards a deleted message's reactions and keeps following
// messages' reactions attached to their new indices.
func (s *Store) RemoveAndShift(message int) {
	if s.items == nil || message < 0 {
		return
	}
	shifted := make(map[int]map[string]bool, len(s.items))
	for index, items := range s.items {
		switch {
		case index < message:
			shifted[index] = items
		case index > message:
			shifted[index-1] = items
		}
	}
	s.items = shifted
}
