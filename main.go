package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/lmittmann/tint"
	"libdb.so/discord-export-messages/internal/flags"
)

var (
	outputMode   = flags.NewStringEnum("md", "json")
	outputFile   = "output"
	mdAllowHTML  = true
	mdDateFormat = "01/02/2006 03:04 PM"
	discordToken = os.Getenv("DISCORD_TOKEN")
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		arg0 := filepath.Base(flag.CommandLine.Name())

		f := func(f string, v ...any) { fmt.Fprintf(flag.CommandLine.Output(), f+"\n", v...) }
		f("Usages:")
		f("  DISCORD_TOKEN= %s [flags] <channelID> <messageID>", arg0)
		f("  DISCORD_TOKEN= %s [flags] <channelID> <messageIDFrom>..<messageIDTo>", arg0)
		f("")
		f("Flags:")
		flag.PrintDefaults()
	}

	flag.Var(outputMode, "output", fmt.Sprintf("output mode (%s)", outputMode.OptionsString()))
	flag.StringVar(&outputFile, "file", outputFile, "output file")
	flag.BoolVar(&mdAllowHTML, "md-allow-html", mdAllowHTML, "allow HTML in markdown output")
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatalf("invalid number of arguments, see -h")
	}

	if discordToken == "" {
		log.Fatalf("missing $DISCORD_TOKEN, see -h")
	}

	logHandler := tint.NewHandler(os.Stderr, nil)
	slog.SetDefault(slog.New(logHandler))

	var out io.Writer
	cleanup := func() {}

	if outputFile == "-" {
		out = os.Stdout
	} else {
		outFile, err := os.Create(outputFile + "." + outputMode.String())
		if err != nil {
			slog.Error(
				"failed to create output file",
				"file", outputFile+"."+outputMode.String(),
				"err", err)
			os.Exit(1)
		}
		defer outFile.Close()

		out = outFile
		cleanup = func() {
			if err := os.Remove(outFile.Name()); err != nil {
				slog.Warn(
					"cannot clean up output file",
					"file", outFile.Name(),
					"err", err)
			}
		}
	}

	if !run(out) {
		cleanup()
		os.Exit(1)
	}
}

func run(w io.Writer) bool {
	channelID := mustSnowflake[discord.ChannelID](flag.Arg(0))

	var messageIDFrom, messageIDTo discord.MessageID
	if strings.Contains(flag.Arg(1), "..") {
		argFrom, argTo, ok := strings.Cut(flag.Arg(1), "..")
		if !ok {
			log.Fatalf("invalid message range: %q", flag.Arg(1))
		}
		messageIDFrom = mustSnowflake[discord.MessageID](argFrom)
		messageIDTo = mustSnowflake[discord.MessageID](argTo)
		if messageIDFrom > messageIDTo {
			// Guarantee that messageIDFrom <= messageIDTo.
			messageIDFrom, messageIDTo = messageIDTo, messageIDFrom
		}
	} else {
		messageIDFrom = mustSnowflake[discord.MessageID](flag.Arg(1))
		messageIDTo = messageIDFrom
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	msgs, err := fetchMessages(ctx, channelID, messageIDFrom, messageIDTo)
	if err != nil {
		slog.Error(
			"failed to fetch messages",
			"err", err)
		return false
	}

	switch outputMode.String() {
	case "md":
		if err := formatMarkdownMessages(ctx, w, msgs); err != nil {
			slog.Error(
				"failed to format messages",
				"output_mode", "md",
				"err", err)
			return false
		}
	case "json":
		if err := formatJSONMessages(ctx, w, msgs); err != nil {
			slog.Error(
				"failed to format messages",
				"output_mode", "json",
				"err", err)
			return false
		}
	}

	return true
}

func fetchMessages(ctx context.Context, chID discord.ChannelID, msgFrom, msgTo discord.MessageID) ([]discord.Message, error) {
	client := api.NewClient(discordToken).WithContext(ctx)
	logger := slog.Default().With(
		"channel_id", chID,
		"message_id_from", msgFrom,
		"message_id_to", msgTo)

	const batch = 100
	allMessages := make([]discord.Message, 0, batch)

	after := msgFrom
	for {
		chunk, err := client.MessagesAfter(chID, after, batch)
		if err != nil {
			return allMessages, fmt.Errorf("failed to fetch chunk after %v: %w", after, err)
		}

		slog.Info(
			"fetched chunk of messages",
			"size", len(chunk),
			"after", after)

		slices.Reverse(chunk)
		allMessages = append(allMessages, chunk...)

		toIx := slices.IndexFunc(chunk, func(m discord.Message) bool { return m.ID >= msgTo })
		if toIx != -1 {
			logger.Info(
				"found message to stop at",
				"size", len(allMessages),
				"chunk_size", len(chunk),
				"chunk_index", toIx)

			allMessages = allMessages[:len(allMessages)-len(chunk)+toIx+1]
			break
		}

		if len(chunk) < batch {
			logger.Info(
				"reached end of messages",
				"size", len(allMessages))
			break
		}

		after = chunk[len(chunk)-1].ID
	}

	return allMessages, nil
}

func mustSnowflake[T ~uint64](s string) T {
	v, err := discord.ParseSnowflake(s)
	if err != nil {
		var x T
		log.Fatalf("invalid %T snowflake: %q", x, s)
	}
	return T(v)
}

func slogFatal(msg string, v ...any) {
	slog.Error(msg, v...)
	log.Fatal(msg)
}
