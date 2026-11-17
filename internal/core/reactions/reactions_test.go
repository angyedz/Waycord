package reactions

import "testing"

func TestRemoveAndShiftKeepsReactionsBoundToMessages(t *testing.T) {
	var store Store
	store.Toggle(0, "❤️")
	store.Toggle(1, "🔥")
	store.Toggle(2, "✅")
	store.Toggle(4, "😄")
	store.RemoveAndShift(1)
	if !store.Has(0, "❤️") || !store.Has(1, "✅") || !store.Has(3, "😄") || store.Has(1, "🔥") {
		t.Fatal("removing a message failed to remove and shift its reaction state")
	}
}
