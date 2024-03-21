package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/diamondburned/arikawa/v3/discord"
)

func formatJSONMessages(_ context.Context, w io.Writer, msgs []discord.Message) error {
	jw := json.NewEncoder(w)
	jw.SetIndent("", "  ")
	return jw.Encode(msgs)
}
