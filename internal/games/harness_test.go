package games_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
	"github.com/Arthur031221/Couchcade/internal/games"
)

type fakeConn struct {
	mu     sync.Mutex
	last   []byte
	all    [][]byte
	closed bool
}

func (f *fakeConn) Send(b []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.last = append([]byte(nil), b...)
	f.all = append(f.all, f.last)
}
func (f *fakeConn) Close() { f.mu.Lock(); f.closed = true; f.mu.Unlock() }

func (f *fakeConn) find(t string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.all) - 1; i >= 0; i-- {
		var m map[string]any
		_ = json.Unmarshal(f.all[i], &m)
		if m["t"] == t {
			return m
		}
	}
	return nil
}

type harness struct {
	t      *testing.T
	m      *engine.Manager
	now    time.Time
	tv     *engine.Session
	tvc    *fakeConn
	code   string
	phones []*engine.Session
	conns  []*fakeConn
}

func newHarness(t *testing.T, players int) *harness {
	t.Helper()
	set, err := content.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, now: time.Unix(1_700_000_000, 0)}
	h.m = engine.NewManager(engine.Config{Content: set, Games: games.New, Catalog: games.Catalog, Seed: 7,
		Clock: func() time.Time { return h.now }})
	h.tvc = &fakeConn{}
	h.tv = engine.NewSession(h.tvc)
	h.m.Handle(h.tv, []byte(`{"t":"tv"}`))
	h.code = h.tvc.find("tvok")["code"].(string)
	for i := 0; i < players; i++ {
		h.join(fmt.Sprintf("P%d", i+1))
	}
	return h
}

func (h *harness) join(name string) int {
	c := &fakeConn{}
	s := engine.NewSession(c)
	h.m.Handle(s, []byte(fmt.Sprintf(`{"t":"join","code":%q,"name":%q}`, h.code, name)))
	h.phones, h.conns = append(h.phones, s), append(h.conns, c)
	return len(h.phones) - 1
}

func (h *harness) host(op string, extra string) {
	msg := fmt.Sprintf(`{"t":"host","op":%q%s}`, op, extra)
	h.m.Handle(h.tv, []byte(msg))
}

func (h *harness) start(game string) {
	h.host("start", fmt.Sprintf(`,"game":%q`, game))
	h.step(0)
}

func (h *harness) act(i int, a engine.Action) {
	b, _ := json.Marshal(map[string]any{"t": "act", "a": a})
	h.m.Handle(h.phones[i], b)
}

// step advances simulated time in 50 ms ticks.
func (h *harness) step(d time.Duration) {
	for el := time.Duration(0); ; el += 50 * time.Millisecond {
		h.m.TickAll()
		if el >= d {
			return
		}
		h.now = h.now.Add(50 * time.Millisecond)
	}
}

// until steps until cond holds or limit passes.
func (h *harness) until(limit time.Duration, cond func() bool) bool {
	for el := time.Duration(0); el < limit; el += 50 * time.Millisecond {
		if cond() {
			return true
		}
		h.step(50 * time.Millisecond)
	}
	return cond()
}

func (h *harness) tvState() (room map[string]any, g map[string]any) {
	m := h.tvc.find("state")
	room, _ = m["room"].(map[string]any)
	g, _ = m["g"].(map[string]any)
	return
}

func (h *harness) phase() string {
	_, g := h.tvState()
	if g == nil {
		return ""
	}
	p, _ := g["phase"].(string)
	return p
}

func (h *harness) roomPhase() string {
	r, _ := h.tvState()
	p, _ := r["phase"].(string)
	return p
}

func (h *harness) score(i int) int {
	r, _ := h.tvState()
	for _, p := range r["players"].([]any) {
		pm := p.(map[string]any)
		if pm["name"] == fmt.Sprintf("P%d", i+1) {
			return int(pm["score"].(float64))
		}
	}
	h.t.Fatalf("player %d not found", i)
	return 0
}

func (h *harness) waitPhase(name string) {
	h.t.Helper()
	if !h.until(3*time.Minute, func() bool { return h.phase() == name }) {
		h.t.Fatalf("never reached phase %q, stuck in %q", name, h.phase())
	}
}

func (h *harness) me(i int) map[string]any {
	m := h.conns[i].find("state")
	me, _ := m["me"].(map[string]any)
	return me
}

func (h *harness) gme(i int) map[string]any {
	g, _ := h.me(i)["g"].(map[string]any)
	return g
}

func (h *harness) errCode(i int) string {
	e := h.conns[i].find("err")
	if e == nil {
		return ""
	}
	return e["code"].(string)
}

func (h *harness) tvg() map[string]any { _, g := h.tvState(); return g }
