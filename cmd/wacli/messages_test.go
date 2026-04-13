package main

import (
	"testing"

	"github.com/steipete/wacli/internal/store"
)

func TestMessageHumanTextPrefersDisplayText(t *testing.T) {
	msg := store.Message{
		Text:        "raw reply",
		DisplayText: "> quoted text\nraw reply",
	}

	if got := messageHumanText(msg); got != "> quoted text\nraw reply" {
		t.Fatalf("expected display text fallback, got %q", got)
	}
	if got := messageRawText(msg); got != "raw reply" {
		t.Fatalf("expected raw text to remain available, got %q", got)
	}
}

func TestMessageHumanTextFallsBackToMediaLabel(t *testing.T) {
	msg := store.Message{MediaType: "image"}

	if got := messageHumanText(msg); got != "Sent image" {
		t.Fatalf("expected media fallback, got %q", got)
	}
	if got := messageRawText(msg); got != "" {
		t.Fatalf("expected no raw text fallback, got %q", got)
	}
}

func TestMessageSenderLabelsPreferName(t *testing.T) {
	msg := store.Message{
		SenderJID:  "123@s.whatsapp.net",
		SenderName: "Alice Example",
	}

	if got := messageSenderLabel(msg); got != "Alice Example" {
		t.Fatalf("expected sender label to prefer name, got %q", got)
	}
	if got := messageSenderDetail(msg); got != "Alice Example (123@s.whatsapp.net)" {
		t.Fatalf("expected sender detail to include name and JID, got %q", got)
	}
}
