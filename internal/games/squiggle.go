package games

import (
	"time"

	"github.com/Arthur031221/Couchcade/internal/engine"
)

const (
	maxStrokes = 60
	maxPoints  = 1500
)

// Squiggle: everyone draws the same prompt, then everyone votes for the best drawing.
type Squiggle struct {
	ph       engine.Phase
	round    int
	rounds   int
	prompts  []string
	drawings map[string][][]int
	order    []string // authors of drawings in the shuffled order they are shown
	votes    map[string]int
	crowd    map[string]int
	gain     map[string]int
	hk       string
	done     bool
}

func (g *Squiggle) Info() engine.Info { return Catalog[0] }

func (g *Squiggle) Start(c *engine.Ctx) {
	g.rounds = c.Rounds(3)
	items := c.Content.Squiggle(c.Lang, c.Family)
	for _, i := range pickN(c.Rng, max(len(items), 1), g.rounds) {
		if len(items) > 0 {
			g.prompts = append(g.prompts, items[i].Text)
		} else {
			g.prompts = append(g.prompts, "?")
		}
	}
	g.round = -1
	g.nextRound(c)
}

func (g *Squiggle) nextRound(c *engine.Ctx) {
	g.round++
	g.drawings, g.votes, g.crowd, g.gain, g.order = map[string][][]int{}, map[string]int{}, map[string]int{}, map[string]int{}, nil
	g.hk = ""
	g.ph.Set("draw", c.Now, 75*time.Second)
}

// sanitize clamps a drawing to the supported grid and size.
func sanitize(p [][]int) [][]int {
	var out [][]int
	total := 0
	for _, s := range p {
		if len(out) >= maxStrokes || len(s) < 4 || len(s)%2 != 0 {
			continue
		}
		if total+len(s)/2 > maxPoints {
			break
		}
		n := []int{min(max(s[0], 0), 11), min(max(s[1], 1), 40)}
		for _, v := range s[2:] {
			n = append(n, min(max(v, 0), 1000))
		}
		total += len(s) / 2
		out = append(out, n)
	}
	return out
}

func (g *Squiggle) Act(c *engine.Ctx, pid string, a engine.Action) string {
	aud := c.IsAudience(pid)
	switch {
	case a.K == "submit" && g.ph.Name == "draw" && !aud:
		d := sanitize(a.P)
		if len(d) == 0 {
			return "empty"
		}
		g.drawings[pid] = d
	case a.K == "vote" && g.ph.Name == "vote":
		if a.I < 0 || a.I >= len(g.order) {
			return ""
		}
		if g.order[a.I] == pid {
			return "own"
		}
		if aud {
			g.crowd[pid] = a.I
		} else {
			g.votes[pid] = a.I
		}
	}
	return ""
}

func (g *Squiggle) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "draw":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.drawings) {
			g.order = shuffled(c.Rng, keys(g.drawings, c.Players()))
			g.hk = "sq-" + string(rune('a'+g.round))
			if len(g.order) < 2 {
				g.tally(c)
				return
			}
			g.ph.Set("vote", c.Now, 30*time.Second)
		}
	case "vote":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.votes) {
			g.tally(c)
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

func (g *Squiggle) tally(c *engine.Ctx) {
	for _, v := range g.votes {
		g.gain[g.order[v]] += 100
	}
	for _, v := range g.crowd {
		g.gain[g.order[v]] += 50
	}
	best := 0
	for _, v := range g.gain {
		best = max(best, v)
	}
	for id, v := range g.gain {
		if v == best && best > 0 {
			g.gain[id] += 100
		}
	}
	if len(g.order) == 1 {
		g.gain[g.order[0]] = 100
	}
	for id, v := range g.gain {
		c.Add(id, v)
	}
	g.ph.Set("reveal", c.Now, 9*time.Second)
}

type sqResult struct {
	Author string `json:"author"`
	Votes  int    `json:"votes"`
	Crowd  int    `json:"crowd"`
	Gain   int    `json:"gain"`
}

func (g *Squiggle) view(c *engine.Ctx) engine.View {
	v := map[string]any{"phase": g.ph.Name, "round": g.round + 1, "rounds": g.rounds, "ms": g.ph.Ms(c.Now), "prompt": g.prompts[g.round], "total": len(c.Active())}
	switch g.ph.Name {
	case "draw":
		v["done"] = keys(g.drawings, c.Players())
	case "vote", "reveal":
		v["voted"] = keys(g.votes, c.Players())
		v["crowd"] = len(g.crowd)
	}
	if g.ph.Name == "reveal" {
		res := make([]sqResult, len(g.order))
		for i, id := range g.order {
			res[i] = sqResult{Author: id, Gain: g.gain[id]}
		}
		for _, i := range g.votes {
			res[i].Votes++
		}
		for _, i := range g.crowd {
			res[i].Crowd++
		}
		v["results"] = res
	}
	if g.done {
		v["phase"] = "end"
	}
	out := engine.View{Shared: v}
	if g.order != nil {
		ds := make([][][]int, len(g.order))
		for i, id := range g.order {
			ds[i] = g.drawings[id]
		}
		out.Heavy, out.HeavyKey = ds, g.hk
	}
	return out
}

func (g *Squiggle) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Squiggle) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Squiggle) Me(c *engine.Ctx, pid string) any {
	me := map[string]any{"submitted": g.drawings[pid] != nil, "own": -1, "vote": -1}
	if i, ok := g.votes[pid]; ok {
		me["vote"] = i
	}
	if i, ok := g.crowd[pid]; ok {
		me["vote"] = i
	}
	for i, id := range g.order {
		if id == pid {
			me["own"] = i
		}
	}
	return me
}

func (g *Squiggle) Done() bool { return g.done }

func (g *Squiggle) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	switch g.ph.Name {
	case "draw":
		if _, ok := g.drawings[pid]; !ok && engine.BotDue(pid, g.ph, c.Now, 6000, 22000) {
			return engine.Action{K: "submit", P: botDoodle(c)}, true
		}
	case "vote":
		if _, ok := g.votes[pid]; !ok && engine.BotDue(pid, g.ph, c.Now, 3000, 12000) && len(g.order) > 1 {
			for {
				if i := c.Rng.Intn(len(g.order)); g.order[i] != pid {
					return engine.Action{K: "vote", I: i}, true
				}
			}
		}
		if _, ok := g.crowd[pid]; !ok && c.IsAudience(pid) && len(g.order) > 1 {
			return engine.Action{K: "vote", I: c.Rng.Intn(len(g.order))}, true
		}
	}
	return engine.Action{}, false
}

// botDoodle draws a few wobbly loops so demo rounds have something to vote on.
func botDoodle(c *engine.Ctx) [][]int {
	var out [][]int
	for s := 0; s < 2+c.Rng.Intn(3); s++ {
		cx, cy, r := 200+c.Rng.Intn(600), 200+c.Rng.Intn(600), 60+c.Rng.Intn(160)
		st := []int{c.Rng.Intn(12), 6 + c.Rng.Intn(10)}
		for a := 0; a <= 24; a++ {
			t := float64(a) / 24 * 6.28318
			w := 1 + 0.25*float64(c.Rng.Intn(100)-50)/50
			x, y := cx+int(float64(r)*w*cosf(t)), cy+int(float64(r)*w*sinf(t))
			st = append(st, min(max(x, 0), 1000), min(max(y, 0), 1000))
		}
		out = append(out, st)
	}
	return out
}
