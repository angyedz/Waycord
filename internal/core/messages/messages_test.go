package messages

import (
	"testing"
	"time"
)

func TestEditOwnTextMessagePreservesOriginalTimestamp(t *testing.T) {
	sentAt := time.Date(2026, 11, 17, 10, 24, 0, 0, time.UTC)
	items := []Message{{Author: "You", Self: true, Body: "before", Time: "10:24", SentAt: sentAt}}
	if !Edit(items, 0, "  after  ") {
		t.Fatal("own text message could not be edited")
	}
	if items[0].Body != "after" || !items[0].Edited || items[0].Time != "10:24" || !items[0].SentAt.Equal(sentAt) {
		t.Fatalf("edit did not preserve its original time: %+v", items[0])
	}
}

func TestEditRejectsOtherAuthorsAndEmptyBody(t *testing.T) {
	items := []Message{{Author: "mira", Body: "hello"}, {Author: "You", Self: true, Body: "hello"}}
	if Edit(items, 0, "changed") || Edit(items, 1, " \n ") || Edit(items, -1, "changed") || Edit(items, 5, "changed") {
		t.Fatal("invalid message edit was accepted")
	}
	if items[0].Body != "hello" || items[1].Body != "hello" {
		t.Fatal("rejected edit changed message content")
	}
}

func TestDeleteRepairsReplyReferencesAndClearsDeletedTarget(t *testing.T) {
	items := []Message{
		{Author: "mira", Body: "source"},
		{Author: "You", Self: true, Body: "delete me"},
		{Author: "You", Self: true, Body: "reply to deleted", HasReply: true, ReplyTo: 1},
		{Author: "mira", Body: "source two"},
		{Author: "You", Self: true, Body: "reply to source two", HasReply: true, ReplyTo: 3},
	}
	backing := items[:cap(items)]
	if Delete(&items, 0) || !Delete(&items, 1) {
		t.Fatal("delete must reject a non-self message and remove an own message")
	}
	if len(items) != 4 || items[1].HasReply || !items[3].HasReply || items[3].ReplyTo != 2 {
		t.Fatalf("reply indices were not repaired: %+v", items)
	}
	if backing[4] != (Message{}) {
		t.Fatalf("deleted slice tail still retains message data: %+v", backing[4])
	}
}
