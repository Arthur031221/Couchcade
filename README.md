<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">Free party games for one TV and everyone's phones. One Go binary, no accounts, no app.</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><b>English</b> | <a href="README.zh-TW.md">Traditional Chinese</a> | <a href="README.zh-CN.md">Simplified Chinese</a> | <a href="README.ja.md">Japanese</a> | <a href="README.ko.md">Korean</a> | <a href="README.es.md">Spanish</a> | <a href="README.fr.md">French</a> | <a href="README.de.md">German</a> | <a href="README.pt-BR.md">Portuguese</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="A Squiggle round on the TV with phones voting"></p>

Open Couchcade on a TV or laptop. Friends scan the QR code, type the four letter code and play from their own phones. There are six original games, an audience mode for crowds, bots for when you're short of people, and nine languages with prompts written for each one.

**Try it now:** the [hosted demo](https://arthur031221.github.io/Couchcade/) runs the same engine compiled to WebAssembly in your browser, with bots filling the seats.

## The number

With 100 simulated phones (8 seated, 92 in the audience), every phone received the first question within 60 ms of the host pressing Start, in 12 out of 12 runs.

How it was measured: `couchcade loadtest -phones 100` starts a server in process, opens 100 real WebSocket connections over loopback, starts a Brain Teams round from the host connection, and timestamps the first state message that carries the question on every phone. Nothing is rendered, so it measures delivery to the phone's socket. The run was repeated 12 times on a Ryzen 5 7500F. Fan-out ranged from 37.5 to 48.1 ms because the server batches changes into one flush every 50 ms, so most of that figure is the batching tick. It measures the server, not your Wi-Fi, which will add its own delay.

## Before and after

| | A typical paid party game box | Couchcade |
| --- | --- | --- |
| Cost | Paid packs | Free, MIT licensed |
| Where it runs | Publisher's apps and servers | One binary on your own network |
| Prompts | Fixed by the publisher | JSON files you can edit and share |
| Languages | Varies by title | 9, with prompts written for each |
| Big groups | Depends on the game | 8 players plus up to 500 audience |

Couchcade isn't a clone of any commercial game. The six games and all prompts are original.

## The games

| Game | What you do |
| --- | --- |
| Squiggle | Draw the prompt in secret, then vote for the best drawing |
| Fibber | Write a fake answer, then spot the real one |
| Brain Teams | Two teams race through multiple choice questions |
| Zap | Wait for green, then tap faster than everyone |
| Echo | Think alike, matching answers score |
| Mole | One player has no secret word, find them |

Rules and player counts are in [docs/how-to-play.md](docs/how-to-play.md).

## Quick start

```sh
couchcade serve
```

Open the first address it prints on the TV. Phones scan the QR code. If your machine has several network cards and the QR code points at the wrong one, pass `-public-url http://192.168.1.20:8080`.

<details>
<summary>Install</summary>

Download a binary from the [releases page](https://github.com/Arthur031221/Couchcade/releases), or:

```sh
# Go 1.24 or newer
go install github.com/Arthur031221/Couchcade/cmd/couchcade@latest

# Docker
docker run --rm -p 8080:8080 ghcr.io/arthur031221/couchcade serve -addr :8080 -public-url http://YOUR-LAN-IP:8080

# Homebrew
brew install Arthur031221/tap/couchcade

# From source
git clone https://github.com/Arthur031221/Couchcade && cd Couchcade
go build -o couchcade ./cmd/couchcade
```

</details>

## Make it yours

Prompts are plain JSON. Drop your own pack in a folder and start with `-packs ./my-packs`:

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` checks a pack and `couchcade packs stats` counts what you have. The format is in [docs/content-packs.md](docs/content-packs.md). Hosting behind a proxy is in [docs/hosting.md](docs/hosting.md) and the design is in [docs/architecture.md](docs/architecture.md).

## What is tested

* Go unit tests for content validation, the engine and every game, run with the race detector in CI.
* A load test with real WebSockets.
* 13 end to end tests that drive real Chrome against the real binary: each of the six games with a human and bots, reconnecting, kicking, more than 8 players, light and dark themes, no sideways scrolling at 320 px in every game, and all nine languages.
* A browser test that plays a full game on the WebAssembly demo.

## Limits

* Phones and the TV need to reach the server, so the usual case is one Wi-Fi network. Playing over the internet needs a reverse proxy you set up.
* Rooms live in memory. Restarting the server ends the games in progress.
* Kick bans a browser token, not a person, and room codes are four letters. Treat both as party tools and not as security.
* There is no login. Don't expose the port to the open internet without a proxy you trust.
* Drawing is finger or mouse only, with undo and clear but no pressure sensitivity.

## Contributing

Prompts and translations are the easiest way in. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)

Assisted by Claude/Codex.
