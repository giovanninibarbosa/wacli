package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/steipete/wacli/internal/out"
	"github.com/steipete/wacli/internal/store"
)

func newMessagesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "List and search messages from the local DB",
	}
	cmd.AddCommand(newMessagesListCmd(flags))
	cmd.AddCommand(newMessagesSearchCmd(flags))
	cmd.AddCommand(newMessagesShowCmd(flags))
	cmd.AddCommand(newMessagesContextCmd(flags))
	return cmd
}

func newMessagesListCmd(flags *rootFlags) *cobra.Command {
	var chat string
	var limit int
	var afterStr string
	var beforeStr string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			var after *time.Time
			var before *time.Time
			if afterStr != "" {
				t, err := parseTime(afterStr)
				if err != nil {
					return err
				}
				after = &t
			}
			if beforeStr != "" {
				t, err := parseTime(beforeStr)
				if err != nil {
					return err
				}
				before = &t
			}

			msgs, err := a.DB().ListMessages(store.ListMessagesParams{
				ChatJID: chat,
				Limit:   limit,
				After:   after,
				Before:  before,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]any{
					"messages": msgs,
					"fts":      a.DB().HasFTS(),
				})
			}

			w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
			fmt.Fprintln(w, "TIME\tCHAT\tFROM\tID\tTEXT")
			for _, m := range msgs {
				from := m.SenderJID
				if m.FromMe {
					from = "me"
				}
				chatLabel := m.ChatName
				if chatLabel == "" {
					chatLabel = m.ChatJID
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					m.Timestamp.Local().Format("2006-01-02 15:04:05"),
					truncate(chatLabel, 24),
					truncate(from, 18),
					truncate(m.MsgID, 14),
					truncate(messageHumanText(m), 80),
				)
			}
			_ = w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&chat, "chat", "", "chat JID")
	cmd.Flags().IntVar(&limit, "limit", 50, "limit results")
	cmd.Flags().StringVar(&afterStr, "after", "", "only messages after time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&beforeStr, "before", "", "only messages before time (RFC3339 or YYYY-MM-DD)")
	return cmd
}

func newMessagesSearchCmd(flags *rootFlags) *cobra.Command {
	var chat string
	var from string
	var limit int
	var afterStr string
	var beforeStr string
	var msgType string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages (FTS5 if available; otherwise LIKE)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			var after *time.Time
			var before *time.Time
			if afterStr != "" {
				t, err := parseTime(afterStr)
				if err != nil {
					return err
				}
				after = &t
			}
			if beforeStr != "" {
				t, err := parseTime(beforeStr)
				if err != nil {
					return err
				}
				before = &t
			}

			msgs, err := a.DB().SearchMessages(store.SearchMessagesParams{
				Query:   args[0],
				ChatJID: chat,
				From:    from,
				Limit:   limit,
				After:   after,
				Before:  before,
				Type:    msgType,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]any{
					"messages": msgs,
					"fts":      a.DB().HasFTS(),
				})
			}

			w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
			fmt.Fprintf(w, "TIME\tCHAT\tFROM\tID\tMATCH\n")
			for _, m := range msgs {
				fromLabel := m.SenderJID
				if m.FromMe {
					fromLabel = "me"
				}
				chatLabel := m.ChatName
				if chatLabel == "" {
					chatLabel = m.ChatJID
				}
				match := m.Snippet
				if match == "" {
					match = messageDisplayText(m)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					m.Timestamp.Local().Format("2006-01-02 15:04:05"),
					truncate(chatLabel, 24),
					truncate(fromLabel, 18),
					truncate(m.MsgID, 14),
					truncate(match, 90),
				)
			}
			_ = w.Flush()
			if !a.DB().HasFTS() {
				fmt.Fprintln(os.Stderr, "Note: FTS5 not enabled; search is using LIKE (slow).")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&chat, "chat", "", "chat JID")
	cmd.Flags().StringVar(&from, "from", "", "sender JID")
	cmd.Flags().IntVar(&limit, "limit", 50, "limit results")
	cmd.Flags().StringVar(&afterStr, "after", "", "only messages after time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&beforeStr, "before", "", "only messages before time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&msgType, "type", "", "media type filter (image|video|audio|document)")
	return cmd
}

func newMessagesShowCmd(flags *rootFlags) *cobra.Command {
	var chat string
	var id string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show one message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if chat == "" || id == "" {
				return fmt.Errorf("--chat and --id are required")
			}

			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			m, err := a.DB().GetMessage(chat, id)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, m)
			}

			fmt.Fprintf(os.Stdout, "Chat: %s\n", m.ChatJID)
			if m.ChatName != "" {
				fmt.Fprintf(os.Stdout, "Chat name: %s\n", m.ChatName)
			}
			fmt.Fprintf(os.Stdout, "ID: %s\n", m.MsgID)
			fmt.Fprintf(os.Stdout, "Time: %s\n", m.Timestamp.Local().Format(time.RFC3339))
			fmt.Fprintf(os.Stdout, "From: %s\n", messageSenderDetail(m))
			if m.MediaType != "" {
				fmt.Fprintf(os.Stdout, "Media: %s\n", m.MediaType)
			}
			if m.MediaCaption != "" {
				fmt.Fprintf(os.Stdout, "Caption: %s\n", m.MediaCaption)
			}
			if m.Filename != "" {
				fmt.Fprintf(os.Stdout, "Filename: %s\n", m.Filename)
			}
			if m.MimeType != "" {
				fmt.Fprintf(os.Stdout, "MIME type: %s\n", m.MimeType)
			}
			if m.LocalPath != "" {
				fmt.Fprintf(os.Stdout, "Downloaded: %s\n", m.LocalPath)
				if !m.DownloadedAt.IsZero() {
					fmt.Fprintf(os.Stdout, "Downloaded at: %s\n", m.DownloadedAt.Local().Format(time.RFC3339))
				}
			}

			fmt.Fprintf(os.Stdout, "\n%s\n", messageHumanText(m))
			if raw := messageRawText(m); raw != "" {
				fmt.Fprintf(os.Stdout, "\nRaw text:\n%s\n", raw)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&chat, "chat", "", "chat JID")
	cmd.Flags().StringVar(&id, "id", "", "message ID")
	return cmd
}

func newMessagesContextCmd(flags *rootFlags) *cobra.Command {
	var chat string
	var id string
	var before int
	var after int

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Show message context around a message ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if chat == "" || id == "" {
				return fmt.Errorf("--chat and --id are required")
			}

			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			msgs, err := a.DB().MessageContext(chat, id, before, after)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, msgs)
			}

			w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
			fmt.Fprintln(w, "TIME\tFROM\tID\tTEXT")
			for _, m := range msgs {
				line := messageHumanText(m)
				if m.MsgID == id {
					line = ">> " + line
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					m.Timestamp.Local().Format("2006-01-02 15:04:05"),
					truncate(messageSenderLabel(m), 18),
					truncate(m.MsgID, 14),
					truncate(line, 100),
				)
			}
			_ = w.Flush()
			return nil
		},
	}
	cmd.Flags().StringVar(&chat, "chat", "", "chat JID")
	cmd.Flags().StringVar(&id, "id", "", "message ID")
	cmd.Flags().IntVar(&before, "before", 5, "messages before")
	cmd.Flags().IntVar(&after, "after", 5, "messages after")
	return cmd
}

func messageDisplayText(m store.Message) string {
	text := strings.TrimSpace(m.DisplayText)
	if text != "" {
		return text
	}
	text = strings.TrimSpace(m.Text)
	if text != "" {
		return text
	}
	if mediaType := strings.TrimSpace(m.MediaType); mediaType != "" {
		return "Sent " + mediaType
	}
	return ""
}

func messageHumanText(m store.Message) string {
	if text := messageDisplayText(m); text != "" {
		return text
	}
	return "(message)"
}

func messageRawText(m store.Message) string {
	raw := strings.TrimSpace(m.Text)
	if raw == "" {
		return ""
	}
	if raw == messageDisplayText(m) {
		return ""
	}
	return raw
}

func messageSenderLabel(m store.Message) string {
	if m.FromMe {
		return "me"
	}
	if name := strings.TrimSpace(m.SenderName); name != "" {
		return name
	}
	return strings.TrimSpace(m.SenderJID)
}

func messageSenderDetail(m store.Message) string {
	if m.FromMe {
		return "me"
	}
	name := strings.TrimSpace(m.SenderName)
	jid := strings.TrimSpace(m.SenderJID)
	switch {
	case name != "" && jid != "" && name != jid:
		return fmt.Sprintf("%s (%s)", name, jid)
	case name != "":
		return name
	case jid != "":
		return jid
	default:
		return "(unknown)"
	}
}
