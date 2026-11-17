package ui

import (
	"testing"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func TestMemberEntriesGroupByServerRole(t *testing.T) {
	members := []mock.Member{
		{Name: "mira", Role: "Members", Status: "Online"},
		{Name: "admin", Role: "Admins", Status: "Online"},
		{Name: "k66", Role: "Members", Status: "Idle"},
		{Name: "bot", Role: "Bots", Bot: true},
	}
	entries := memberEntries(members)
	if len(entries) != 7 {
		t.Fatalf("expected three role headers and four members, got %d entries", len(entries))
	}
	if entries[0].role != "Members" || entries[0].count != 2 || entries[1].member.Name != "mira" || entries[2].member.Name != "k66" {
		t.Fatalf("member order or first role group changed: %+v", entries[:3])
	}
	if entries[3].role != "Admins" || entries[3].count != 1 || entries[5].role != "Bots" || entries[5].count != 1 || entries[6].member.Name != "bot" {
		t.Fatalf("unexpected role groups: %+v", entries)
	}
}

func TestMemberEntriesOmitEmptyGroups(t *testing.T) {
	if entries := memberEntries(nil); len(entries) != 0 {
		t.Fatalf("empty server produced member groups: %+v", entries)
	}
}
