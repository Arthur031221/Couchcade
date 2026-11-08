# Architecture

```
TV page, phones or the WASM demo
        |  WebSocket (or a JS callback)
     server  ->  engine  ->  games (squiggle, fibber, trivia, zap, echo, mole)
                    |
                 content (JSON packs)
```

## Engine

`internal/engine` knows nothing about HTTP. It talks to a `Conn` with a non blocking `Send`. A `Manager` owns rooms, a `Room` owns players and one `Game`, and the room lock serializes every change.

State is pushed, not polled. Mutations set a dirty flag and a 50 ms tick flushes one `state` message per connection. Large data such as drawings travels in a `heavy` field that is only resent when its key `hk` changes.

Bots are ordinary players whose actions come from `BotDue`, so the same code paths run with and without humans.

## Games

Each game implements `engine.Game`: start, receive an action, tick on the clock, and render a view per player. They are tested with a fake clock and fake connections, so a whole game runs in milliseconds.

## Server

`internal/server` adapts WebSockets to `Conn`, serves the embedded pages, draws QR codes at `/api/qr`, and reports `/healthz`. A connection whose outgoing queue fills up is dropped. The phone reconnects with its token.

## Web

`web/` is plain ES modules with no build step. A small keyed virtual DOM (`h`, `patch`) renders each game as a `tv(S, C)` and a `phone(S, C)` function. Strings come from `locales/<lang>.json` with an English fallback.

## Hosted demo

`cmd/couchcade-wasm` compiles the same engine to WebAssembly and exposes `__couchEngine.open(callback)` to the page. The demo page puts a TV and phone frames side by side, each speaking the same wire protocol as a real WebSocket.

## Wire protocol

Phone or TV to server:

| Message | Fields |
| --- | --- |
| `tv` | `code`, `token` (empty to open a new room) |
| `join` | `code`, `name`, `token` |
| `act` | `a: {k, s, i, p}` an action for the current game |
| `host` | `op` of start, lobby, lang, family, kick, mute, bots, nobots |
| `ping` | `n` |

Server to client: `tvok`, `joined`, `state`, `err`, `kicked`, `pong`. Error codes are `room_not_found`, `name_required`, `room_full`, `kicked`, `bad_op`, `busy`, `need_players`, `unknown_game`, `empty`, `own`, `real_answer` and `secret_word`.

## Tests

* `go test ./...` covers content validation, every game and the engine.
* `couchcade loadtest` opens real WebSockets against an in process server.
* `e2e/run.mjs` drives real Chrome against the real binary through all six games, reconnects, kicks, audience, themes, 320 px layouts and all nine languages.
* `e2e/demo.mjs` plays a game on the WASM demo.
