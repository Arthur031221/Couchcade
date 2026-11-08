<p align="center"><img src="assets/logo.svg" width="96" alt="Couchcade logo"></p>
<h1 align="center">Couchcade</h1>
<p align="center">Kostenlose Partyspiele für einen Fernseher und die Handys aller. Eine einzige Go-Datei, ohne Konto und ohne App.</p>

<p align="center">
<a href="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml"><img src="https://github.com/Arthur031221/Couchcade/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT"></a>
</p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-TW.md">繁體中文</a> | <a href="README.zh-CN.md">简体中文</a> | <a href="README.ja.md">日本語</a> | <a href="README.ko.md">한국어</a> | <a href="README.es.md">Español</a> | <a href="README.fr.md">Français</a> | <b>Deutsch</b> | <a href="README.pt-BR.md">Português</a></p>

<p align="center"><img src="assets/demo.gif" width="760" alt="Eine Squiggle-Runde auf dem Fernseher, Handys stimmen ab"></p>

Couchcade auf dem Fernseher oder Laptop öffnen. Freunde scannen den QR-Code, tippen den Code aus vier Buchstaben ein und spielen mit dem eigenen Handy. Es gibt sechs eigene Spiele, einen Zuschauermodus für große Runden, Bots für den Fall, dass Leute fehlen, und neun Sprachen mit Aufgaben, die für jede Sprache eigens geschrieben wurden. Die Punkte laufen über alle Spiele hinweg in eine Gesamtwertung des Abends, und die Handys bleiben beim Spielen wach.

**Gleich ausprobieren:** Die [Online-Demo](https://arthur031221.github.io/Couchcade/) lässt dieselbe Engine als WebAssembly im Browser laufen, Bots besetzen die Plätze.

## Die Zahl

Mit 100 simulierten Handys (8 Sitzplätze, 92 Zuschauer) erhielt jedes Handy die erste Frage innerhalb von 60 ms, nachdem der Host auf Start gedrückt hatte, in 12 von 12 Läufen.

So wurde gemessen: `couchcade loadtest -phones 100` startet einen Server im selben Prozess, öffnet 100 echte WebSocket-Verbindungen über Loopback, startet von der Host-Verbindung aus eine Runde Brain Teams und stoppt, wann die erste Statusnachricht mit der Frage auf jedem Handy ankommt. Der Lauf wurde 12 Mal auf einem Ryzen 5 7500F wiederholt. Die Verteilung dauerte 37,5 bis 48,1 ms, weil der Server Änderungen bündelt und alle 50 ms sendet, der größte Teil dieser Zahl ist also der Bündelungstakt. Gemessen wird der Server, nicht dein WLAN, das seine eigene Verzögerung dazugibt. Es wird nichts gezeichnet, gemessen wird also die Zustellung bis zum Socket des Handys.

## Vorher und nachher

|  | Ein typisches kostenpflichtiges Partyspiel | Couchcade |
| --- | --- | --- |
| Kosten | Bezahlte Pakete | Kostenlos, MIT-Lizenz |
| Wo es läuft | Apps und Server des Herausgebers | Eine Datei im eigenen Netzwerk |
| Inhalte | Vom Herausgeber festgelegt | JSON-Dateien zum Bearbeiten und Teilen |
| Sprachen | Je nach Titel | 9, mit eigens geschriebenen Aufgaben |
| Große Gruppen | Je nach Spiel | 8 Spielende und bis zu 500 Zuschauer |

Couchcade ist kein Klon eines kommerziellen Spiels. Die sechs Spiele und alle Aufgaben sind eigene Werke.

## Die Spiele

| Spiel | Was man tut |
| --- | --- |
| Squiggle | Das Thema heimlich zeichnen, dann die beste Zeichnung wählen |
| Fibber | Eine falsche Antwort schreiben, dann die echte erkennen |
| Brain Teams | Zwei Teams rasen durch Multiple-Choice-Fragen |
| Zap | Auf Grün warten und schneller tippen als alle anderen |
| Echo | Gleich denken, übereinstimmende Antworten punkten |
| Mole | Eine Person kennt das geheime Wort nicht, finde sie |

Regeln und Spielerzahlen stehen in [docs/how-to-play.md](docs/how-to-play.md).

## Schnellstart

```sh
couchcade serve
```

Die erste ausgegebene Adresse am Fernseher öffnen, die Handys scannen den QR-Code. Hat der Rechner mehrere Netzwerkkarten und der QR-Code zeigt auf die falsche, hilft `-public-url http://192.168.1.20:8080`.

<details>
<summary>Installation</summary>

Eine Datei von der [Releases-Seite](https://github.com/Arthur031221/Couchcade/releases) laden oder:

```sh
# Go 1.24+
go install github.com/Arthur031221/Couchcade/cmd/couchcade@latest

# Docker
docker build -t couchcade https://github.com/Arthur031221/Couchcade.git
docker run --rm -p 8080:8080 couchcade serve -addr :8080 -public-url http://YOUR-LAN-IP:8080

# Source
git clone https://github.com/Arthur031221/Couchcade && cd Couchcade
go build -o couchcade ./cmd/couchcade
```

</details>

## Eigene Inhalte

Die Aufgaben sind einfaches JSON. Ein eigenes Paket in einen Ordner legen und mit `-packs ./my-packs` starten:

```json
{"id": "office", "lang": "en", "items": {"squiggle": [{"text": "a tired printer"}]}}
```

`couchcade packs validate FILE` prüft ein Paket und `couchcade packs stats` zählt, was vorhanden ist. Das Format steht in [docs/content-packs.md](docs/content-packs.md), der Betrieb hinter einem Proxy in [docs/hosting.md](docs/hosting.md) und der Aufbau in [docs/architecture.md](docs/architecture.md).

## Was getestet wird

* Go-Unit-Tests für Inhaltsprüfung, Engine und jedes Spiel, in der CI mit Race Detector.
* Ein Lasttest mit echten WebSockets.
* 16 Ende-zu-Ende-Tests, die ein echtes Chrome gegen die echte Datei steuern: jedes der sechs Spiele mit Mensch und Bots, Wiederverbinden, Rauswerfen, mehr als 8 Spielende, helles und dunkles Design, kein seitliches Scrollen bei 320 px in jedem Spiel und alle neun Sprachen mit übersetzten Schaltflächenbeschriftungen.
* Ein Browsertest, der auf der WebAssembly-Demo ein ganzes Spiel spielt.

## Grenzen

* Handys und Fernseher müssen den Server erreichen, der Normalfall ist also ein gemeinsames WLAN. Übers Internet zu spielen braucht einen Reverse Proxy, den du selbst einrichtest.
* Räume liegen im Arbeitsspeicher. Ein Neustart des Servers beendet laufende Spiele.
* Es gibt keine Anmeldung. Den Port nicht ohne vertrauenswürdigen Proxy ins offene Internet stellen.
* Gezeichnet wird nur mit Finger oder Maus, mit Rückgängig und Löschen, aber ohne Druckempfindlichkeit.

## Mitmachen

Aufgaben und Übersetzungen sind der einfachste Einstieg. Siehe [CONTRIBUTING.md](CONTRIBUTING.md).

## Lizenz

[MIT](LICENSE)

Assisted by Claude/Codex.
