# ALIAS's status reporter office bot

## What is this?
An internal binding between [simple-status-reporter](https://git.sr.ht/~alias/simple-status-reporter) and Discord for ALIAS, the CS department's student association at Sorbonne Université, Paris

## Staff-facing usage
This is a Go program, so a [Go toolchain](https://go.dev/) is required

Make a new Discord bot on their [developer portal](https://discord.com/developers/applications), obtain their token, and find out where your `simple-status-reporter` client certificate and private key are.

You have to create the toml config file (default path `/etc/office-bot/config.toml`) with the following values:

```toml
bot-token = ""
cert-path = ""
private-key-path = ""
```

If you want to use a different path, you can specify it with the `-config` flag.

**This is currently a work in progress!** Contributions are welcome in any case after discussion on [our Discord server](https://discord.gg/Qq6u8Mz)
