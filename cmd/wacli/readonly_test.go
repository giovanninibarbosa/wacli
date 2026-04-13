package main

import "testing"

func TestSendTextBlockedInReadOnlyMode(t *testing.T) {
	flags := &rootFlags{readOnly: true}
	cmd := newSendTextCmd(flags)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--to", "1234567890", "--message", "hello"})

	err := cmd.Execute()
	if err == nil || err.Error() != readOnlyErrorMessage {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestAuthBlockedInReadOnlyMode(t *testing.T) {
	flags := &rootFlags{readOnly: true}
	cmd := newAuthCmd(flags)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil || err.Error() != readOnlyErrorMessage {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestGroupsRenameBlockedInReadOnlyMode(t *testing.T) {
	flags := &rootFlags{readOnly: true}
	cmd := newGroupsRenameCmd(flags)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--jid", "123456@g.us", "--name", "Renamed"})

	err := cmd.Execute()
	if err == nil || err.Error() != readOnlyErrorMessage {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestContactsAliasSetBlockedInReadOnlyMode(t *testing.T) {
	flags := &rootFlags{readOnly: true}
	cmd := newContactsAliasCmd(flags)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"set", "--jid", "1234567890@s.whatsapp.net", "--alias", "Alice"})

	err := cmd.Execute()
	if err == nil || err.Error() != readOnlyErrorMessage {
		t.Fatalf("expected read-only error, got %v", err)
	}
}
