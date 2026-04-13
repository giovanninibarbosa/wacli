//go:build !sqlite_fts5

package store

import (
	"testing"
	"time"
)

func TestSearchMessagesUsesLIKEWhenFTSDisabled(t *testing.T) {
	db := openTestDB(t)
	if db.HasFTS() {
		t.Fatalf("expected HasFTS=false in !sqlite_fts5 build")
	}

	chat := "123@s.whatsapp.net"
	if err := db.UpsertChat(chat, "dm", "Alice", time.Now()); err != nil {
		t.Fatalf("UpsertChat: %v", err)
	}
	if err := db.UpsertMessage(UpsertMessageParams{
		ChatJID:    chat,
		ChatName:   "Alice",
		MsgID:      "m1",
		SenderJID:  chat,
		SenderName: "Alice",
		Timestamp:  time.Now(),
		FromMe:     false,
		Text:       "hello world",
	}); err != nil {
		t.Fatalf("UpsertMessage: %v", err)
	}

	ms, err := db.SearchMessages(SearchMessagesParams{Query: "hello", Limit: 10})
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(ms) != 1 {
		t.Fatalf("expected 1 result, got %d", len(ms))
	}
	if ms[0].Snippet != "" {
		t.Fatalf("expected empty snippet for LIKE search, got %q", ms[0].Snippet)
	}
}

func TestSearchMessagesAppliesMediaFiltersWithoutFTS(t *testing.T) {
	db := openTestDB(t)
	chat := seedSearchFilterMessages(t, db)

	withMedia, err := db.SearchMessages(SearchMessagesParams{
		Query:    "project",
		ChatJID:  chat,
		HasMedia: true,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("SearchMessages with media: %v", err)
	}
	if len(withMedia) != 2 {
		t.Fatalf("expected 2 media results, got %d", len(withMedia))
	}
	for _, msg := range withMedia {
		if msg.MediaType == "" {
			t.Fatalf("expected only media results, got %+v", msg)
		}
	}

	textOnly, err := db.SearchMessages(SearchMessagesParams{
		Query:   "project",
		ChatJID: chat,
		Type:    "text",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("SearchMessages text: %v", err)
	}
	if len(textOnly) != 1 || textOnly[0].MsgID != "m-text" {
		t.Fatalf("expected only text result, got %+v", textOnly)
	}

	imageOnly, err := db.SearchMessages(SearchMessagesParams{
		Query:   "project",
		ChatJID: chat,
		Type:    "image",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("SearchMessages image: %v", err)
	}
	if len(imageOnly) != 1 || imageOnly[0].MsgID != "m-image" {
		t.Fatalf("expected only image result, got %+v", imageOnly)
	}
}
