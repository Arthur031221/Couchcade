package engine

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ"

// Config wires a Manager to its content and games.
type Config struct {
	Content    *content.Set
	Games      func(id string) Game
	Catalog    []Info
	MaxPlayers int
	// JoinURL builds the address a phone opens for a room code.
	JoinURL func(code string) string
	Clock   func() time.Time
	Seed    int64
}

// Manager owns every room and routes messages to them.
type Manager struct {
	cfg   Config
	mu    sync.Mutex
	rooms map[string]*Room
	rnd   *rand.Rand
}

func NewManager(cfg Config) *Manager {
	if cfg.MaxPlayers <= 0 {
		cfg.MaxPlayers = 8
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	if cfg.JoinURL == nil {
		cfg.JoinURL = func(code string) string { return "/play?c=" + code }
	}
	return &Manager{cfg: cfg, rooms: map[string]*Room{}, rnd: rand.New(rand.NewSource(cfg.Seed))}
}

func (m *Manager) joinURL(code string) string { return m.cfg.JoinURL(code) }

func (m *Manager) langs() []string { return content.Languages }

func token(r *rand.Rand) string {
	b := make([]byte, 12)
	if _, err := crand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(r.Intn(256))
		}
	}
	return hex.EncodeToString(b)
}

// Session is one connection's place in a room.
type Session struct {
	conn     Conn
	room     *Room
	pid      string
	role     string
	heavyKey string
	tokens   float64
	last     time.Time
	closed   atomic.Bool
}

func NewSession(c Conn) *Session { return &Session{conn: c, tokens: 60, last: time.Now()} }

func (s *Session) send(b []byte) {
	if !s.closed.Load() {
		s.conn.Send(b)
	}
}

func (s *Session) sendState(room, shared []byte, hk string, heavy, me []byte) {
	if hk == s.heavyKey {
		heavy = nil
	} else {
		s.heavyKey = hk
	}
	s.send(buildState(room, shared, hk, heavy, me))
}

func (s *Session) detach() {
	s.closed.Store(true)
	s.conn.Close()
}

// allow is a token bucket: 30 messages a second with a burst of 60.
func (s *Session) allow() bool {
	now := time.Now()
	s.tokens += now.Sub(s.last).Seconds() * 30
	if s.tokens > 60 {
		s.tokens = 60
	}
	s.last = now
	if s.tokens < 1 {
		return false
	}
	s.tokens--
	return true
}

func errMsg(code string) []byte { return []byte(`{"t":"err","code":"` + code + `"}`) }

// maxRooms bounds memory when someone opens rooms in a loop.
const maxRooms = 1000

// newCode picks a free four letter code, or "" when it keeps colliding.
func (m *Manager) newCode() string {
	for try := 0; try < 200; try++ {
		b := make([]byte, 4)
		for i := range b {
			b[i] = codeAlphabet[m.rnd.Intn(len(codeAlphabet))]
		}
		if c := string(b); m.rooms[c] == nil {
			return c
		}
	}
	return ""
}

type inbound struct {
	T     string          `json:"t"`
	Code  string          `json:"code"`
	Name  string          `json:"name"`
	Token string          `json:"token"`
	N     json.Number     `json:"n"`
	Op    string          `json:"op"`
	A     Action          `json:"a"`
	Raw   json.RawMessage `json:"-"`
}

// Handle processes one message from a connection.
func (m *Manager) Handle(s *Session, raw []byte) {
	if !s.allow() {
		return
	}
	var in inbound
	if json.Unmarshal(raw, &in) != nil {
		return
	}
	switch in.T {
	case "ping":
		s.send([]byte(`{"t":"pong","n":` + string(in.N) + `}`))
	case "tv":
		m.attachTV(s, strings.ToUpper(in.Code), in.Token)
	case "join":
		m.attachPhone(s, strings.ToUpper(strings.TrimSpace(in.Code)), in.Name, in.Token)
	case "act":
		if r := s.room; r != nil {
			r.mu.Lock()
			r.lastActive = r.now
			if code := r.act(s.pid, in.A); code != "" {
				s.send(errMsg(code))
			}
			r.mu.Unlock()
		}
	case "host":
		r := s.room
		if r == nil {
			return
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		r.mu.Lock()
		if s.role == "tv" || s.pid == r.hostID() {
			if code := r.host(in.Op, fields); code != "" {
				s.send(errMsg(code))
			}
		}
		r.mu.Unlock()
	}
}

func (m *Manager) attachTV(s *Session, code, tok string) {
	m.mu.Lock()
	r := m.rooms[code]
	if r == nil || r.tvToken != tok {
		if len(m.rooms) >= maxRooms {
			m.mu.Unlock()
			s.send(errMsg("busy"))
			return
		}
		code = m.newCode()
		if code == "" {
			m.mu.Unlock()
			s.send(errMsg("busy"))
			return
		}
		r = newRoom(m, code, m.cfg.Clock())
		m.rooms[code] = r
	}
	old := s.room
	m.mu.Unlock()
	if old != nil && old != r {
		// One connection owns one room: leaving the old one lets it be swept.
		old.mu.Lock()
		old.leave(s)
		empty := old.tv == nil && len(old.players) == 0
		old.mu.Unlock()
		if empty {
			m.mu.Lock()
			delete(m.rooms, old.code)
			m.mu.Unlock()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tv != nil && r.tv != s {
		r.tv.detach()
	}
	r.tv, s.room, s.role = s, r, "tv"
	r.dirty = true
	s.send([]byte(`{"t":"tvok","code":"` + r.code + `","token":"` + r.tvToken + `"}`))
	r.flush()
}

func (m *Manager) attachPhone(s *Session, code, name, tok string) {
	m.mu.Lock()
	r := m.rooms[code]
	m.mu.Unlock()
	if r == nil {
		s.send(errMsg("room_not_found"))
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.room != nil {
		return // already seated: a second join would leave a ghost player behind
	}
	p, e := r.join(s, name, tok)
	if e != "" {
		s.send(errMsg(e))
		return
	}
	s.send([]byte(`{"t":"joined","id":"` + p.ID + `","token":"` + p.token + `","code":"` + r.code + `"}`))
	r.flush()
}

// Close is called when a connection ends.
func (m *Manager) Close(s *Session) {
	s.closed.Store(true)
	if s.room != nil {
		s.room.mu.Lock()
		s.room.leave(s)
		s.room.mu.Unlock()
	}
}

// Run ticks every room until ctx ends and removes abandoned rooms.
func (m *Manager) Run(ctx context.Context) {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	lastSweep := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			m.TickAll()
			if now.Sub(lastSweep) > time.Minute {
				m.Sweep()
				lastSweep = now
			}
		}
	}
}

// TickAll advances every room to the configured clock's time.
func (m *Manager) TickAll() {
	m.mu.Lock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	m.mu.Unlock()
	now := m.cfg.Clock()
	for _, r := range rooms {
		r.Tick(now)
	}
}

// Sweep drops rooms whose TV left more than 30 minutes ago.
func (m *Manager) Sweep() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for code, r := range m.rooms {
		r.mu.Lock()
		gone := r.tv == nil && !r.anyoneConnected() && r.now.Sub(r.lastActive) > 30*time.Minute
		r.mu.Unlock()
		if gone {
			delete(m.rooms, code)
		}
	}
}

// RoomCount is used by the health endpoint.
func (m *Manager) RoomCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rooms)
}
