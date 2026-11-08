# Hosting

## On your own network

```sh
couchcade serve
```

The program prints two addresses. Open the first on the TV and let phones use the second. Phones and the TV must be on the same network. Couchcade guesses the first private IPv4 address of the machine. If it guesses wrong, say which one phones should use:

```sh
couchcade serve -public-url http://192.168.1.20:8080
```

The QR code on the TV encodes that address.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-addr` | `:8080` | Listen address |
| `-public-url` | LAN address | Address phones open |
| `-packs` | none | Folder of extra content packs |
| `-max-players` | 8 | Seats per room, the rest become audience |
| `-allow-origin` | none | Comma separated hosts allowed to open WebSockets behind a proxy |
| `-time-scale` | 1 | Run game clocks faster, used by the tests only |

## Docker

```sh
docker build -t couchcade https://github.com/Arthur031221/Couchcade.git
docker run --rm -p 8080:8080 couchcade
```

Inside a container the LAN address is not visible, so pass `-public-url`:

```sh
docker run --rm -p 8080:8080 couchcade serve -addr :8080 -public-url http://192.168.1.20:8080
```

## Behind a reverse proxy

Terminate TLS in the proxy and forward `/` including WebSocket upgrades on `/ws`. Set `-public-url https://games.example.com` and, if the proxy rewrites the `Host` header, `-allow-origin games.example.com`.

Caddy:

```
games.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

## Limits worth knowing

* One process keeps every room in memory. Restarting it ends all games.
* Rooms are removed 30 minutes after their TV disconnects.
* There is no login. Anyone who can reach the server can open a room, so do not expose it to the internet without a proxy that you trust.
* Each connection may send 30 messages per second.
* Measured on one machine: `couchcade loadtest -phones 100` (see the README for the numbers).
* Kicking bans a browser token, not a person. Someone who opens a private tab can rejoin, so treat kick as a way to clear the room and not as security.
* Room codes are four letters. Anyone on the network can try codes, so keep the server off networks you do not trust.
