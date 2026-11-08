# Content packs

Prompts live in JSON files. The nine built in packs are in `internal/content/packs`. Add your own with `-packs DIR`: every `*.json` file in the folder is merged with the built in prompts for its language.

```json
{
  "id": "my-pack",
  "lang": "en",
  "name": "Office night",
  "author": "you",
  "items": {
    "squiggle": [{"text": "a tired printer"}],
    "fibber":   [{"q": "The first email was sent in ___.", "a": "1971", "rating": "family"}],
    "trivia":   [{"q": "Capital of Canada?", "options": ["Toronto", "Ottawa", "Montreal", "Quebec"], "answer": 1}],
    "echo":     [{"text": "Name a tool", "hints": ["hammer", "saw"]}],
    "mole":     [{"category": "Office", "word": "stapler"}]
  }
}
```

## Rules

* `id` is required and `lang` must be one of `en zh-TW zh-CN ja ko es fr de pt-BR`.
* `rating` is optional, either `family` or `mature`. Items without a rating count as family.
* Squiggle text is up to 60 characters, echo text up to 80.
* Fibber needs a question of up to 200 characters and an answer of up to 40.
* Trivia needs distinct options and `answer` is the zero based index of the right one.
* `hints` are typical answers that bots use. They are optional.
* Duplicates inside one pack are refused.
* Unknown fields are refused so typos do not pass silently.

## Check a pack

```sh
couchcade packs validate my-pack.json
couchcade packs stats -packs ./my-packs
```

The validator prints every problem it finds. `stats` prints how many prompts each language has per game.
