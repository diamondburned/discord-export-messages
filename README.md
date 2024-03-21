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

## Example Output

> <img src="https://cdn.discordapp.com/avatars/170132746042081280/164568b5bb1bff0906728e5274d0e18a.png?size=48" alt="Diamond's avatar" width="20" /> **Diamond** – *03/21 08:33 AM*
> 
> real
> 
> plan to out all the EU people

> <img src="https://cdn.discordapp.com/avatars/170939974227591168/a_3c70dbb3d44e875938c4384d228a8e75.png?size=48" alt="topi's avatar" width="20" /> **topi** – *03/21 08:33 AM*
> 
> <img src="https://cdn.discordapp.com/emojis/1029896994631716905.gif?size=96&amp;quality=lossless" alt=":BocchiHide:" width="48" />

```sh
# Ran with:
discord-export-messages --md-allow-html ${CHANNEL_ID} ${MESSAGE_ID[0]}..${MESSAGE_ID[1]}
```
