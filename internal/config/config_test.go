package config

import "testing"

func TestReadOnlyEnabled(t *testing.T) {
	t.Setenv("WACLI_READONLY", "1")

	if !ReadOnlyEnabled() {
		t.Fatalf("expected WACLI_READONLY=1 to enable read-only mode")
	}
}

func TestReadOnlyEnabledFalseByDefault(t *testing.T) {
	t.Setenv("WACLI_READONLY", "")

	if ReadOnlyEnabled() {
		t.Fatalf("expected empty WACLI_READONLY to leave read-only mode disabled")
	}
}
