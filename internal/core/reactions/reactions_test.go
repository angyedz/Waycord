package reactions

import "testing"

func TestToggleKeepsMessagesIndependent(t *testing.T) {
	var s Store
	s.Toggle(2, "👍")
	s.Toggle(3, "👍")
	s.Toggle(2, "🔥")
	s.Toggle(2, "👍")
	if s.Has(2, "👍") || !s.Has(3, "👍") || !s.Has(2, "🔥") {
		t.Fatal("reactions were not isolated or toggled")
	}
}
