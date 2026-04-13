package main

import "testing"

func TestNormalizeMessageType(t *testing.T) {
	tests := map[string]string{
		"":         "",
		"text":     "text",
		" IMAGE ":  "image",
		"video":    "video",
		"audio":    "audio",
		"document": "document",
		"gif":      "gif",
		"sticker":  "sticker",
	}

	for input, want := range tests {
		got, err := normalizeMessageType(input)
		if err != nil {
			t.Fatalf("normalizeMessageType(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizeMessageType(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestNormalizeMessageTypeRejectsUnsupportedValues(t *testing.T) {
	if _, err := normalizeMessageType("voice-note"); err == nil {
		t.Fatalf("expected unsupported type to return an error")
	}
}
