package games

import (
	"sort"
	"time"

	"github.com/Arthur031221/Couchcade/internal/engine"
)

var zapPoints = []int{300, 200, 150, 100, 75, 50, 50, 50}

// Zap: tap when the screen flashes. Tapping early costs points.
type Zap struct {
	ph     engine.Phase
	round  int
	rounds int
	goAt   time.Time
	taps   map[string]time.Time
	early  map[string]bool
	gain   map[string]int
	ranks  []zapRank
	done   bool
}

type zapRank struct {
	ID string `json:"id"`
	Ms int    `json:"ms"`
}

func (g *Zap) Info() engine.Info { return Catalog[3] }

func (g *Zap) Start(c *engine.Ctx) {
	g.rounds = c.Rounds(5)
	g.round = -1
	g.nextRound(c)
}

func (g *Zap) nextRound(c *engine.Ctx) {
	g.round++
	g.taps, g.early, g.gain, g.ranks = map[string]time.Time{}, map[string]bool{}, map[string]int{}, nil
	wait := 2*time.Second + time.Duration(c.Rng.Intn(4000))*time.Millisecond
	g.ph.Set("wait", c.Now, wait)
	g.goAt = g.ph.Until
}

func (g *Zap) Act(c *engine.Ctx, pid string, a engine.Action) string {
	if a.K != "tap" || c.IsAudience(pid) {
		return ""
	}
	if _, tapped := g.taps[pid]; tapped || g.early[pid] {
		return ""
	}
	phase := g.ph.Name
	if phase == "wait" && !c.Now.Before(g.goAt) {
		phase = "go"
	}
	switch phase {
	case "wait":
		g.early[pid] = true
		g.gain[pid] = -100
		c.Add(pid, -100)
	case "go":
		g.taps[pid] = c.Now
	}
	return ""
}

func (g *Zap) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "wait":
		if g.ph.Over(c.Now) {
			g.ph.Set("go", g.goAt, 3*time.Second)
		}
	case "go":
		if g.ph.Over(c.Now) || allActed(c.Active(), merged(g.taps, g.early)) {
			g.score(c)
			g.ph.Set("reveal", c.Now, 6*time.Second)
		}
	case "reveal":
		if g.ph.Over(c.Now) {
			if g.round+1 >= g.rounds {
				g.done = true
				return
			}
			g.nextRound(c)
		}
	}
}

func merged(taps map[string]time.Time, early map[string]bool) map[string]bool {
	m := map[string]bool{}
	for id := range taps {
		m[id] = true
	}
	for id := range early {
		m[id] = true
	}
	return m
}

func (g *Zap) score(c *engine.Ctx) {
	for id, t := range g.taps {
		g.ranks = append(g.ranks, zapRank{id, int(t.Sub(g.goAt).Milliseconds())})
	}
	sort.Slice(g.ranks, func(i, j int) bool { return g.ranks[i].Ms < g.ranks[j].Ms })
	for i, r := range g.ranks {
		p := zapPoints[len(zapPoints)-1]
		if i < len(zapPoints) {
			p = zapPoints[i]
		}
		g.gain[r.ID] = p
		c.Add(r.ID, p)
	}
}

func (g *Zap) view(c *engine.Ctx) engine.View {
	v := map[string]any{"phase": g.ph.Name, "round": g.round + 1, "rounds": g.rounds, "total": len(c.Active())}
	switch g.ph.Name {
	case "wait":
		v["early"] = keys(g.early, c.Players())
	case "go":
		v["tapped"] = len(g.taps)
	case "reveal":
		v["ranks"], v["early"], v["gain"] = g.ranks, keys(g.early, c.Players()), g.gain
	}
	if g.done {
		v["phase"] = "end"
	}
	return engine.View{Shared: v}
}

func (g *Zap) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Zap) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Zap) Me(c *engine.Ctx, pid string) any {
	_, tapped := g.taps[pid]
	return map[string]any{"tapped": tapped, "early": g.early[pid], "gain": g.gain[pid]}
}

func (g *Zap) Done() bool { return g.done }

func (g *Zap) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	if g.ph.Name != "go" {
		return engine.Action{}, false
	}
	if _, tapped := g.taps[pid]; tapped || !engine.BotDue(pid, g.ph, c.Now, 260, 900) {
		return engine.Action{}, false
	}
	return engine.Action{K: "tap"}, true
}
