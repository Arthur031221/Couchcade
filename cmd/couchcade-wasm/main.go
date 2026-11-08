//go:build js && wasm

// Command couchcade-wasm runs the whole game engine inside the browser so the
// hosted demo needs no server. The page opens "sockets" through
// window.__couchEngine.open and speaks the same JSON protocol as the real one.
package main

import (
	"context"
	"sync"
	"syscall/js"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
	"github.com/Arthur031221/Couchcade/internal/games"
)

// conn delivers engine output to a JS callback from its own goroutine, so the
// engine never calls into JavaScript while holding a room lock.
type conn struct {
	out  chan []byte
	in   chan []byte
	done chan struct{}
	once sync.Once
}

func (c *conn) Send(b []byte) {
	select {
	case c.out <- append([]byte(nil), b...):
	default: // a stalled page drops a frame instead of blocking the room
	}
}

func (c *conn) Close() {
	c.once.Do(func() { close(c.done) })
}

func main() {
	set, err := content.Builtin()
	if err != nil {
		panic(err)
	}
	mgr := engine.NewManager(engine.Config{
		Content: set, Games: games.New, Catalog: games.Catalog, MaxPlayers: 8,
		JoinURL: func(code string) string { return "play.html?c=" + code },
	})
	go mgr.Run(context.Background())

	open := js.FuncOf(func(_ js.Value, args []js.Value) any {
		cb := args[0]
		c := &conn{out: make(chan []byte, 64), in: make(chan []byte, 64), done: make(chan struct{})}
		s := engine.NewSession(c)
		go func() {
			for {
				select {
				case b := <-c.out:
					cb.Invoke(string(b))
				case <-c.done:
					return
				}
			}
		}()
		go func() {
			for {
				select {
				case b := <-c.in:
					mgr.Handle(s, b)
				case <-c.done:
					mgr.Close(s)
					return
				}
			}
		}()
		sock := js.Global().Get("Object").New()
		sock.Set("send", js.FuncOf(func(_ js.Value, a []js.Value) any {
			select {
			case c.in <- []byte(a[0].String()):
			default:
			}
			return nil
		}))
		sock.Set("close", js.FuncOf(func(js.Value, []js.Value) any { c.Close(); return nil }))
		return sock
	})
	api := js.Global().Get("Object").New()
	api.Set("open", open)
	js.Global().Set("__couchEngine", api)
	if ready := js.Global().Get("__couchReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {}
}
