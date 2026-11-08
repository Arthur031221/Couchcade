# Contributing

Thanks for helping. The fastest ways to help are prompts and translations, which need no Go.

## Add or fix prompts

Edit a pack in `internal/content/packs` or send a new one, then run:

```sh
go run ./cmd/couchcade packs validate internal/content/packs/*.json
go test ./internal/content
```

See `docs/content-packs.md` for the format. Keep prompts short and something a mixed group can say aloud. Mark anything edgy `"rating": "mature"`.

## Add a language

1. Copy `web/locales/en.json` to `web/locales/<code>.json` and translate every value. Keep `{placeholders}` as they are.
2. Add a pack `internal/content/packs/<code>.json` with the same item counts as the English one.
3. Add the code to `Languages` in `internal/content/content.go` and to the language list at the top of `web/js/i18n.js`.

`go test ./...` fails when a locale misses a key or a language has too few prompts.

## Code

```sh
go vet ./...
go test -race ./...
go build -o couchcade ./cmd/couchcade
cd e2e && npm install && node run.mjs ../couchcade
```

The e2e run needs Chrome. Set `CHROME` if it is not at `/usr/bin/google-chrome`.

Rules of the road:

* Engine and games stay free of HTTP and of wall clock calls. Take the time from the room.
* A new game is a Go file in `internal/games`, a JS file in `web/js/games`, strings in every locale and a test that plays it through with bots.
* The frontend has no build step. Do not add one.
* Keep the diff small and say what you ran.
