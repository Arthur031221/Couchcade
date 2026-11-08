package engine

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	PhaseLobby = "lobby"
	PhaseGame  = "game"
	PhaseFinal = "final"

	finalHold    = 30 * time.Second
	maxAudience  = 500
	reconnectTTL = 10 * time.Minute
)

type Player struct {
	ID       string
	Name     string
	Color    int
	Score    int
	Total    int // points from finished games, kept until the host resets standings
	Bot      bool
	Muted    bool
	Audience bool
	token    string
	sess     *Session
	seen     time.Time
}

// Room is one TV with its phones.
type Room struct {
	mu         sync.Mutex
	m          *Manager
	code       string
	now        time.Time
	rng        *rand.Rand
	lang       string
	family     bool
	rounds     int
	played     int // games finished since the last reset
	maxPlayers int
	players    map[string]*Player
	order      []string
	tv         *Session
	tvToken    string
	phase      string
	gameID     string
	game       Game
	finalAt    time.Time
	dirty      bool
	lastActive time.Time
	nextID     int
	banned     map[string]bool
}

func newRoom(m *Manager, code string, now time.Time) *Room {
	return &Room{
		m: m, code: code, now: now, rng: rand.New(rand.NewSource(m.cfg.Seed + int64(len(m.rooms)))),
		lang: "en", family: true, maxPlayers: m.cfg.MaxPlayers, players: map[string]*Player{},
		phase: PhaseLobby, lastActive: now, banned: map[string]bool{}, tvToken: token(m.rnd),
	}
}

func (r *Room) ctx() *Ctx {
	return &Ctx{Now: r.now, Rng: r.rng, Lang: r.lang, Family: r.family, Content: r.m.cfg.Content, room: r}
}

func (r *Room) seated() []string {
	out := make([]string, 0, len(r.order))
	for _, id := range r.order {
		if !r.players[id].Audience {
			out = append(out, id)
		}
	}
	return out
}

func (r *Room) audienceCount() int {
	n := 0
	for _, p := range r.players {
		if p.Audience && (p.sess != nil || p.Bot) {
			n++
		}
	}
	return n
}

func (r *Room) anyoneConnected() bool {
	for _, p := range r.players {
		if p.sess != nil {
			return true
		}
	}
	return false
}

func (r *Room) hostID() string {
	fallback := ""
	for _, id := range r.seated() {
		p := r.players[id]
		if p.Bot {
			continue
		}
		if p.sess != nil {
			return id
		}
		if fallback == "" {
			fallback = id
		}
	}
	return fallback
}

func cleanName(s string) string {
	var b strings.Builder
	for _, ru := range strings.TrimSpace(s) {
		if unicode.IsControl(ru) || unicode.Is(unicode.Cf, ru) || ru == '<' || ru == '>' {
			continue
		}
		b.WriteRune(ru)
	}
	rs := []rune(strings.TrimSpace(b.String()))
	if len(rs) > 16 {
		rs = rs[:16]
	}
	return string(rs)
}

func (r *Room) uniqueName(name string) string {
	taken := func(n string) bool {
		for _, p := range r.players {
			if strings.EqualFold(p.Name, n) {
				return true
			}
		}
		return false
	}
	if !taken(name) {
		return name
	}
	for i := 2; ; i++ {
		if n := name + " " + strconv.Itoa(i); !taken(n) {
			return n
		}
	}
}

func (r *Room) addPlayer(name string, bot bool) *Player {
	r.nextID++
	p := &Player{ID: "p" + strconv.Itoa(r.nextID), Name: r.uniqueName(name), Color: (r.nextID - 1) % 12, Bot: bot, seen: r.now}
	p.token = token(r.m.rnd)
	if r.phase != PhaseLobby || len(r.seated()) >= r.maxPlayers {
		p.Audience = true
	}
	r.players[p.ID] = p
	r.order = append(r.order, p.ID)
	return p
}

// join attaches a session as a player or audience member. It returns the
// player, or nil with an error code.
func (r *Room) join(s *Session, name, tok string) (*Player, string) {
	if tok != "" {
		if r.banned[tok] {
			return nil, "kicked"
		}
		for _, p := range r.players {
			if p.token == tok && !p.Bot {
				if p.sess != nil && p.sess != s {
					p.sess.detach()
				}
				p.sess, p.seen = s, r.now
				s.room, s.pid, s.role = r, p.ID, "phone"
				r.dirty = true
				return p, ""
			}
		}
	}
	name = cleanName(name)
	if name == "" {
		return nil, "name_required"
	}
	if r.audienceCount() >= maxAudience || len(r.players) >= maxAudience+2*r.m.cfg.MaxPlayers {
		return nil, "room_full"
	}
	p := r.addPlayer(name, false)
	p.sess = s
	s.room, s.pid, s.role = r, p.ID, "phone"
	r.dirty = true
	return p, ""
}

func (r *Room) leave(s *Session) {
	if s.role == "tv" && r.tv == s {
		r.tv, r.lastActive = nil, r.now
	}
	if p := r.players[s.pid]; p != nil && p.sess == s {
		p.sess, p.seen = nil, r.now
	}
	r.dirty = true
}

func (r *Room) promoteAudience() {
	for _, id := range r.order {
		p := r.players[id]
		if p.Audience && len(r.seated()) < r.maxPlayers && (p.Bot || p.sess != nil) {
			p.Audience = false
		}
	}
}

func (r *Room) startGame(id string, rounds int) string {
	if r.phase != PhaseLobby && r.phase != PhaseFinal {
		return "busy"
	}
	if r.phase == PhaseFinal {
		r.promoteAudience() // "Play again" skips the lobby, so seat waiting audience first
	}
	g := r.m.cfg.Games(id)
	if g == nil {
		return "unknown_game"
	}
	n := 0
	for _, pid := range r.seated() {
		if p := r.players[pid]; p.Bot || p.sess != nil {
			n++
		}
	}
	if n < g.Info().Min {
		return "need_players"
	}
	if rounds < 0 || rounds > 12 {
		rounds = 0
	}
	r.rounds = rounds
	for _, p := range r.players {
		p.Score = 0
	}
	r.game, r.gameID, r.phase = g, id, PhaseGame
	g.Start(r.ctx())
	r.dirty = true
	return ""
}

func (r *Room) backToLobby() {
	r.phase, r.game, r.gameID = PhaseLobby, nil, ""
	r.promoteAudience()
	r.dirty = true
}

// host applies a host operation from the TV or the host phone.
func (r *Room) host(op string, raw map[string]json.RawMessage) string {
	str := func(k string) string {
		var v string
		_ = json.Unmarshal(raw[k], &v)
		return v
	}
	num := func(k string) int {
		var v int
		_ = json.Unmarshal(raw[k], &v)
		return v
	}
	switch op {
	case "start":
		return r.startGame(str("game"), num("rounds"))
	case "lobby":
		r.backToLobby()
	case "lang":
		l := str("lang")
		for _, k := range r.m.langs() {
			if k == l {
				r.lang = l
				r.dirty = true
			}
		}
	case "resettotals":
		if r.phase != PhaseLobby {
			return "busy"
		}
		r.played = 0
		for _, p := range r.players {
			p.Total = 0
		}
		r.dirty = true
	case "family":
		var v bool
		_ = json.Unmarshal(raw["on"], &v)
		r.family = v
		r.dirty = true
	case "kick":
		if p := r.players[str("id")]; p != nil {
			r.banned[p.token] = true
			if p.sess != nil {
				p.sess.send([]byte(`{"t":"kicked"}`))
				p.sess.detach()
			}
			delete(r.players, p.ID)
			for i, id := range r.order {
				if id == p.ID {
					r.order = append(r.order[:i], r.order[i+1:]...)
					break
				}
			}
			r.dirty = true
		}
	case "mute":
		if p := r.players[str("id")]; p != nil {
			p.Muted = !p.Muted
			r.dirty = true
		}
	case "bots":
		if r.phase != PhaseLobby {
			return "busy"
		}
		for i := 0; i < num("n") && i < 12 && len(r.seated()) < r.maxPlayers; i++ {
			r.addPlayer(botNames[r.rng.Intn(len(botNames))], true)
		}
		r.dirty = true
	case "nobots":
		if r.phase != PhaseLobby {
			return "busy"
		}
		for _, id := range append([]string(nil), r.order...) {
			if r.players[id].Bot {
				delete(r.players, id)
				for i, o := range r.order {
					if o == id {
						r.order = append(r.order[:i], r.order[i+1:]...)
						break
					}
				}
			}
		}
		r.dirty = true
	default:
		return "bad_op"
	}
	return ""
}

var botNames = []string{"Pixel", "Mochi", "Waffle", "Comet", "Biscuit", "Nova", "Pudding", "Zigzag", "Pickle", "Turbo", "Maple", "Otter"}

func (r *Room) act(pid string, a Action) string {
	p := r.players[pid]
	if p == nil || r.game == nil || r.phase != PhaseGame || p.Muted {
		return ""
	}
	if t := r.m.cfg.Clock(); t.After(r.now) {
		r.now = t
	}
	r.lastActive = r.now
	r.dirty = true
	return r.game.Act(r.ctx(), pid, a)
}

// Tick advances the clock: phase deadlines, bots, reconnect expiry, flush.
func (r *Room) Tick(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.now = now
	switch r.phase {
	case PhaseGame:
		c := r.ctx()
		r.game.Tick(c)
		for _, id := range r.seated() {
			if p := r.players[id]; p.Bot && !p.Muted {
				if a, ok := r.game.Bot(c, id); ok {
					r.game.Act(c, id, a)
				}
			}
		}
		for _, id := range r.order {
			if p := r.players[id]; p.Audience && p.Bot {
				if a, ok := r.game.Bot(c, id); ok {
					r.game.Act(c, id, a)
				}
			}
		}
		r.dirty = true
		if r.game.Done() {
			r.phase, r.finalAt = PhaseFinal, now.Add(finalHold)
			r.played++
			for _, p := range r.players {
				p.Total += p.Score
			}
		}
	case PhaseFinal:
		if !now.Before(r.finalAt) {
			r.backToLobby()
		}
	}
	for _, id := range append([]string(nil), r.order...) {
		p := r.players[id]
		if !p.Bot && p.sess == nil && now.Sub(p.seen) > reconnectTTL {
			delete(r.players, id)
			for i, o := range r.order {
				if o == id {
					r.order = append(r.order[:i], r.order[i+1:]...)
					break
				}
			}
		}
	}
	if r.dirty {
		r.flush()
	}
}

// PlayerView is the public part of a player.
type PlayerView struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color int    `json:"color"`
	Score int    `json:"score"`
	Total int    `json:"total"`
	On    bool   `json:"on"`
	Bot   bool   `json:"bot,omitempty"`
	Muted bool   `json:"muted,omitempty"`
}

type roomView struct {
	Code     string       `json:"code"`
	Lang     string       `json:"lang"`
	Family   bool         `json:"family"`
	Phase    string       `json:"phase"`
	Game     string       `json:"game,omitempty"`
	Rounds   int          `json:"rounds"`
	Played   int          `json:"played"`
	Max      int          `json:"max"`
	Host     string       `json:"host"`
	Join     string       `json:"join"`
	Audience int          `json:"audience"`
	Players  []PlayerView `json:"players"`
	Catalog  []Info       `json:"catalog,omitempty"`
	Langs    []string     `json:"langs,omitempty"`
}

func (r *Room) roomJSON() []byte {
	v := roomView{Code: r.code, Lang: r.lang, Family: r.family, Phase: r.phase, Game: r.gameID, Rounds: r.rounds, Played: r.played,
		Max: r.maxPlayers, Host: r.hostID(), Join: r.m.joinURL(r.code), Audience: r.audienceCount(), Players: []PlayerView{}}
	for _, id := range r.seated() {
		p := r.players[id]
		v.Players = append(v.Players, PlayerView{p.ID, p.Name, p.Color, p.Score, p.Total, p.sess != nil || p.Bot, p.Bot, p.Muted})
	}
	if r.phase == PhaseLobby {
		v.Catalog, v.Langs = r.m.cfg.Catalog, r.m.langs()
	}
	b, _ := json.Marshal(v)
	return b
}

func marshal(v any) []byte {
	if v == nil {
		return []byte("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}
	return b
}

// flush sends the current state to every connection.
func (r *Room) flush() {
	r.dirty = false
	rb := r.roomJSON()
	var tvv, phv View
	var c *Ctx
	if r.game != nil {
		c = r.ctx()
		tvv, phv = r.game.TVView(c), r.game.PhoneView(c)
	}
	tvS, phS := marshal(tvv.Shared), marshal(phv.Shared)
	tvH, phH := heavyJSON(tvv), heavyJSON(phv)
	if r.tv != nil {
		r.tv.sendState(rb, tvS, tvv.HeavyKey, tvH, nil)
	}
	for _, id := range r.order {
		p := r.players[id]
		if p.sess == nil {
			continue
		}
		me := map[string]any{"id": id, "aud": p.Audience, "host": id == r.hostID(), "score": p.Score, "total": p.Total, "muted": p.Muted}
		if c != nil {
			me["g"] = r.game.Me(c, id)
		}
		p.sess.sendState(rb, phS, phv.HeavyKey, phH, marshal(me))
	}
}

func heavyJSON(v View) []byte {
	if v.Heavy == nil {
		return nil
	}
	return marshal(v.Heavy)
}

func buildState(room, shared []byte, hk string, heavy, me []byte) []byte {
	var b bytes.Buffer
	b.WriteString(`{"t":"state","room":`)
	b.Write(room)
	b.WriteString(`,"g":`)
	b.Write(shared)
	if hk != "" {
		b.WriteString(`,"hk":`)
		b.Write(marshal(hk))
		if heavy != nil {
			b.WriteString(`,"heavy":`)
			b.Write(heavy)
		}
	}
	if me != nil {
		b.WriteString(`,"me":`)
		b.Write(me)
	}
	b.WriteByte('}')
	return b.Bytes()
}
