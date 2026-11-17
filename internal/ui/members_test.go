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

func TestMemberEntriesShowActiveActivitiesBeforeRoleGroups(t *testing.T) {
	members := []mock.Member{
		{Name: "mira", Role: "Members", Status: "Online", Activity: "Listening to music"},
		{Name: "k66", Role: "Members", Status: "Idle", Activity: "Playing a game"},
		{Name: "offline", Role: "Members", Status: "Offline", Activity: "Old activity"},
		{Name: "plain", Role: "Members", Status: "Online"},
	}
	entries := memberEntries(members)
	if len(entries) != 8 {
		t.Fatalf("expected one activity header, two activity cards, one role header, and four members; got %+v", entries)
	}
	if !entries[0].activity || entries[0].member != nil || entries[0].count != 2 {
		t.Fatalf("activity header should count only active members with an activity: %+v", entries[0])
	}
	if entries[1].member.Name != "mira" || !entries[1].activity || entries[2].member.Name != "k66" || !entries[2].activity {
		t.Fatalf("activity cards were missing or out of order: %+v", entries[:3])
	}
	if entries[3].role != "Members" || entries[3].activity || entries[4].member.Name != "mira" || entries[7].member.Name != "plain" {
		t.Fatalf("role members should follow the activity section: %+v", entries[3:])
	}
}
