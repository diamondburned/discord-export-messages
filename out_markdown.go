package main

import (
	"bufio"
	"context"
	"fmt"
	"go/doc"
	"go/doc/comment"
	"html"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

type markdownContentKey struct {
	id      discord.MessageID
	content string
}

func formatMarkdownMessages(_ context.Context, w io.Writer, msgs []discord.Message) error {
	msgs = slices.Clone(msgs)
	slices.SortFunc(msgs, func(a, b discord.Message) int {
		// Sort earliest to latest.
		return int(a.ID - b.ID)
	})

	bw := bufio.NewWriter(w)

	renderContentCache := make(map[markdownContentKey]string)
	renderContent := func(msg discord.Message, content string) string {
		cacheKey := markdownContentKey{msg.ID, content}
		if content, ok := renderContentCache[cacheKey]; ok {
			return content
		}

		content = strings.ReplaceAll(content, "\n", "\n\n")
		content = renderEmojis(content)
		content = renderMentions(content, msgs)
		content = renderTenorLinks(content, msg)
		renderContentCache[cacheKey] = content

		return content
	}

	for i, msg := range msgs {
		collapse := i > 0 && msg.Author.ID == msgs[i-1].Author.ID
		if !collapse {
			if i > 0 {
				fmt.Fprintf(bw, "\n")
			}

			fmt.Fprintf(bw, "> ")
			if mdAllowHTML {
				fmt.Fprintf(bw,
					`<img src="%s?size=48" alt="%s's avatar" width="20" />`,
					html.EscapeString(msg.Author.AvatarURLWithType(discord.PNGImage)),
					html.EscapeString(msg.Author.DisplayOrUsername()))
			} else {
				fmt.Fprintf(bw,
					"![%s's avatar](%s?size=20)",
					msg.Author.DisplayOrUsername(),
					msg.Author.AvatarURLWithType(discord.PNGImage))
			}

			fmt.Fprintf(bw,
				" **%s** – *%s*\n",
				msg.Author.DisplayOrUsername(),
				msg.Timestamp.Format(mdDateFormat))
		}

		fmt.Fprintf(bw, "> \n")

		if msg.Reference != nil {
			if msg.ReferencedMessage == nil {
				fmt.Fprintf(bw, "> > Replying to *unknown message*.\n")
				fmt.Fprintf(bw, "> \n")
			} else {
				fmt.Fprintf(bw, "> > Replying to **%s**: %s\n",
					msg.ReferencedMessage.Author.DisplayOrUsername(),
					renderContent(msg, singleLine(msg.ReferencedMessage.Content)))
				fmt.Fprintf(bw, "> \n")
			}
		}

		fmt.Fprintf(bw, "%s\n", mdQuote(renderContent(msg, msg.Content)))
	}

	return bw.Flush()
}

func singleLine(text string) string {
	if strings.Contains(text, "\n") {
		text, _, _ = strings.Cut(text, "\n")
	}
	text = wrapText(text, 80)
	if strings.Contains(text, "\n") {
		text, _, _ = strings.Cut(text, "\n")
		text += "..."
	}
	return text
}

func wrapText(text string, col int) string {
	d := new(doc.Package).Parser().Parse(text)
	pr := &comment.Printer{
		TextPrefix:     "",
		TextCodePrefix: "  ",
		TextWidth:      col,
	}
	return string(pr.Text(d))
}

var reEmoji = regexp.MustCompile(`<(a)?:([^:]+):(\d+)>`)

func mdQuote(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return strings.Join(lines, "\n")
}

func renderEmojis(text string) string {
	size := 24
	if strings.TrimSpace(reEmoji.ReplaceAllString(text, "")) == "" {
		size = 48
	}

	return reEmoji.ReplaceAllStringFunc(text, func(match string) string {
		submatches := reEmoji.FindStringSubmatch(match)

		animated := submatches[1] == "a"
		name := submatches[2]
		id := submatches[3]

		ext := "webp"
		if animated {
			ext = "gif"
		}

		var markup string
		if mdAllowHTML {
			url := fmt.Sprintf(
				"https://cdn.discordapp.com/emojis/%s.%s?size=%d&quality=lossless",
				id, ext, size*2)

			markup = fmt.Sprintf(
				"<img src=\"%s\" alt=\":%s:\" width=\"%d\" />",
				html.EscapeString(url),
				html.EscapeString(name),
				size)
		} else {
			url := fmt.Sprintf(
				"https://cdn.discordapp.com/emojis/%s.%s?size=%d&quality=lossless",
				id, ext, size)

			markup = fmt.Sprintf(
				"![:%s:](%s)",
				name, url)
		}

		return markup
	})
}

var reMention = regexp.MustCompile(`<@!?(\d+)>`)

func renderMentions(text string, messages []discord.Message) string {
	knownAuthors := make(map[discord.UserID]string)
	for _, msg := range messages {
		knownAuthors[msg.Author.ID] = msg.Author.DisplayOrUsername()
	}

	return reMention.ReplaceAllStringFunc(text, func(match string) string {
		submatches := reMention.FindStringSubmatch(match)
		id := mustSnowflake[discord.UserID](submatches[1])

		name, ok := knownAuthors[id]
		if !ok {
			name = "[*unknown*]"
		}

		return fmt.Sprintf("@%s", name)
	})
}

var reTenorLink = regexp.MustCompile(`https://tenor.com/view/(\S+)`)

func renderTenorLinks(text string, msg discord.Message) string {
	return reTenorLink.ReplaceAllStringFunc(text, func(match string) string {
		ix := slices.IndexFunc(msg.Embeds, func(e discord.Embed) bool { return e.URL == match })
		if ix == -1 {
			slog.Warn(
				"failed to find Tenor GIF embed",
				"message_id", msg.ID,
				"tenor_url", match)
			return match
		}

		thumbnail := msg.Embeds[ix].Thumbnail
		if thumbnail == nil {
			slog.Warn(
				"Tenor GIF embed has no thumbnail",
				"message_id", msg.ID,
				"tenor_url", match)
			return match
		}

		if mdAllowHTML {
			return fmt.Sprintf(
				`<a href="%s"><img alt="Tenor GIF" src="%s" width="400px" /></a>`,
				html.EscapeString(match),
				html.EscapeString(thumbnail.URL))
		} else {
			return fmt.Sprintf(
				"[%s](![Tenor GIF](%s))",
				match, thumbnail.URL)
		}
	})
}
