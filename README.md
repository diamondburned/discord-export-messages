# discord-export-messages

Tool to export Discord messages to a file in Markdown or JSON format. It is
intended to be shared via GitHub Gists.

## Usage

```
Usages:
  DISCORD_TOKEN= discord-export-messages [flags] <channelID> <messageID>
  DISCORD_TOKEN= discord-export-messages [flags] <channelID> <messageIDFrom>..<messageIDTo>

Flags:
  -file string
    	output file (default "output")
  -md-allow-html
    	allow HTML in markdown output (default true)
  -output value
    	output mode (md, json) (default md)
```

> [!NOTE]
> The `DISCORD_TOKEN` environment variable must be set with a valid Discord token
> to use the tool.

> [!WARNING]
> Using the tool with a user token is against the Discord Terms of Service.
> You may be banned if you use it with a user token. Use at your own risk.
