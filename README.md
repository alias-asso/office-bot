# ALIAS's status reporter office bot

## What is this?
A simple discord bot to set the online status info of [ALIAS](https://alias-asso.fr) office in Sorbonne Université, Paris.

## Staff-facing usage
This is a Go program, so a [Go toolchain](https://go.dev/) is required

Make a new Discord bot on their [developer portal](https://discord.com/developers/applications) and obtain its token.

Create a webhook in the channel of your choice and note its id and token.

You have to create the toml config file (default path `/etc/office-bot/config.toml`) with the following values:

```toml
bot-token = ""
status-file-path = ""
server-port = ""
password-hash = ""
toggle-roles = [ # a list of discord roles IDs
  "", # role1
	"",  # role2
  ""   # ...
]
webhook-id = ""
webhook-token = ""
```

If you want to use a different path, you can specify it with the `-config` flag.

**This is currently a work in progress!** Contributions are welcome in any case after discussion on [our Discord server](https://discord.gg/Qq6u8Mz)
