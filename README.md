# Botsu

Botsu is a Discord bot for logging time spent on immersion activities in Japanese.
It is the successor to [AnotherImmersionBot](https://github.com/UTD-JLA/another-immersion-bot),
which itself is an inspiration of TheMoeWay's [ImmersionBot](https://github.com/TheMoeWay/immersion-bot).

## Requirements

- PostgreSQL 15
- Go 1.21 (for running from source)
- Docker (for running from Docker image)

## Install
To install the latest release from the terminal, use:
```sh
curl -L https://github.com/UTD-JLA/botsu/releases/latest/download/botsu-linux-amd64 -o ~/.local/bin/botsu
chmod a+rx ~/.local/bin/botsu
```
replacing `~/.local/bin` with your preferred install location. 

To install from source, use:
```
go install github.com/UTD-JLA/botsu/cmd/botsu@latest
```
which will install to your `GOBIN` (`~/go/bin` by default).

## Basic Setup

1. Create a Discord bot account and invite it to your server.
2. Create a new PostgreSQL database, for example `botsu`.
3. Clone this repository.
4. Create a `config.toml` file in your working directory with the following contents:
```toml
token = "your bot token"
# This requires the privileged members intent, which can be enabled in the Discord developer portal.
use_members_intent = true

[database]
user = "your database user"
password = "your database password"
host = "your database host"
port = 5432
database = "botsu"
```
5. Run `botsu` from the working directory. Assuming you are working from the repository root:
```bash
go run ./cmd/botsu -migrations=file://./migrations
```

You will not need to run the migrations again unless there is a change to the database schema,
but it will not hurt to pass the argument on every run.

## Keeping data sources up to date
The first time Botsu runs, it will need to download some data sources which may take a while.
The default location for these files is `./data`, and you will also find `*-index.bluge` files in your
working directory. These files can be safely deleted, but Botsu will need to re-download them the next time it runs.
This may be desired when you want to update the data sources.
