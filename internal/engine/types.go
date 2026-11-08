// Package engine runs rooms: players, reconnects, the phase clock and bots.
// It knows nothing about sockets, so the same code serves the HTTP server and
// the WebAssembly demo.
package engine

import (
	"math/rand"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
)

// Conn is one live connection. Send must never block.
type Conn interface {
	Send(data []byte)
	Close()
}

// Action is what a phone sends during a game. Games read the fields they need.
type Action struct {
	K string  `json:"k"`
	S string  `json:"s,omitempty"`
	I int     `json:"i,omitempty"`
	P [][]int `json:"p,omitempty"`
}

// Info describes a game for the lobby.
type Info struct {
	ID     string `json:"id"`
	Min    int    `json:"min"`
	Rounds int    `json:"rounds"`
}

// View is what a game exposes to one class of viewer. Heavy data is resent
// only when HeavyKey changes, so large payloads such as drawings are not
// repeated on every vote.
type View struct {
	Shared   any
	Heavy    any
	HeavyKey string
}

// Game is a state machine. All methods run under the room lock.
type Game interface {
	Info() Info
	Start(c *Ctx)
	// Act handles a phone action. A non-empty error code is shown to that phone.
	Act(c *Ctx, pid string, a Action) string
	Tick(c *Ctx)
	TVView(c *Ctx) View
	PhoneView(c *Ctx) View
	// Me is the small per-player part of a phone view.
	Me(c *Ctx, pid string) any
	Done() bool
	// Bot returns the next action for a simulated phone, if it has one.
	Bot(c *Ctx, pid string) (Action, bool)
}

// Ctx is the view of the room a game is allowed to use.
type Ctx struct {
	Now     time.Time
	Rng     *rand.Rand
	Lang    string
	Family  bool
	Content *content.Set
	room    *Room
}

// Players are the seated players in join order.
func (c *Ctx) Players() []string { return c.room.seated() }

// Active are seated players who can still act: connected or bots, not muted.
func (c *Ctx) Active() []string {
	var out []string
	for _, id := range c.room.seated() {
		p := c.room.players[id]
		if (p.Bot || p.sess != nil) && !p.Muted {
			out = append(out, id)
		}
	}
	return out
}

func (c *Ctx) IsAudience(id string) bool {
	p := c.room.players[id]
	return p != nil && p.Audience
}

func (c *Ctx) Name(id string) string {
	if p := c.room.players[id]; p != nil {
		return p.Name
	}
	return ""
}

func (c *Ctx) Add(id string, n int) {
	if p := c.room.players[id]; p != nil {
		p.Score += n
	}
}

func (c *Ctx) Score(id string) int {
	if p := c.room.players[id]; p != nil {
		return p.Score
	}
	return 0
}

func (c *Ctx) AudienceCount() int { return c.room.audienceCount() }

// Rounds returns the host's choice or def when the host left it on default.
func (c *Ctx) Rounds(def int) int {
	if c.room.rounds > 0 {
		return c.room.rounds
	}
	return def
}

// Phase is a named stage that ends at a deadline.
type Phase struct {
	Name  string
	Start time.Time
	Until time.Time
}

func (p *Phase) Set(name string, now time.Time, d time.Duration) {
	p.Name, p.Start, p.Until = name, now, now.Add(d)
}

func (p *Phase) Over(now time.Time) bool { return !now.Before(p.Until) }

// Ms is the time left, for clients that draw a countdown.
func (p *Phase) Ms(now time.Time) int {
	d := p.Until.Sub(now).Milliseconds()
	if d < 0 {
		return 0
	}
	return int(d)
}

// BotDue reports whether a simulated phone should act now. The delay is a
// stable pseudo-random value per player and phase, so bots do not all act in
// the same tick.
func BotDue(pid string, p Phase, now time.Time, loMs, hiMs int) bool {
	h := uint32(2166136261)
	for _, b := range []byte(pid) {
		h = (h ^ uint32(b)) * 16777619
	}
	h = (h ^ uint32(p.Start.UnixNano()>>20)) * 16777619
	span := hiMs - loMs
	if span < 1 {
		span = 1
	}
	return now.Sub(p.Start) >= time.Duration(loMs+int(h%uint32(span)))*time.Millisecond
}
