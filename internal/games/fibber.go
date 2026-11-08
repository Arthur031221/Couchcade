package games

import (
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
)

type fibOption struct {
	Text   string
	Owners []string // empty for the real answer
	Real   bool
}

// Fibber: write a believable fake answer, then find the real one among the fakes.
type Fibber struct {
	ph      engine.Phase
	round   int
	rounds  int
	qs      []content.QA
	pool    []string // other questions' answers, used as bot fakes and fillers
	fakes   map[string]string
	options []fibOption
	picks   map[string]int
	crowd   map[string]int // audience picks, worth less and never decisive
	gain    map[string]int
	done    bool
}

func (g *Fibber) Info() engine.Info { return Catalog[1] }

func (g *Fibber) Start(c *engine.Ctx) {
	g.rounds = c.Rounds(3)
	items := c.Content.Fibber(c.Lang, c.Family)
	for _, i := range pickN(c.Rng, max(len(items), 1), g.rounds) {
		if len(items) > 0 {
			g.qs = append(g.qs, items[i])
		} else {
			g.qs = append(g.qs, content.QA{Q: "?", A: "?"})
		}
	}
	for _, it := range items {
		g.pool = append(g.pool, it.A)
	}
	g.round = -1
	g.nextRound(c)
}

func (g *Fibber) nextRound(c *engine.Ctx) {
	g.round++
	g.fakes, g.picks, g.crowd, g.gain, g.options = map[string]string{}, map[string]int{}, map[string]int{}, map[string]int{}, nil
	g.ph.Set("write", c.Now, 50*time.Second)
}

func (g *Fibber) Act(c *engine.Ctx, pid string, a engine.Action) string {
	aud := c.IsAudience(pid)
	switch {
	case a.K == "submit" && g.ph.Name == "write":
		if aud {
			return ""
		}
		t := clean(a.S, 40)
		if norm(t) == "" {
			return "empty"
		}
		if norm(t) == norm(g.qs[g.round].A) {
			return "real_answer"
		}
		g.fakes[pid] = t
	case a.K == "pick" && g.ph.Name == "pick":
		if a.I < 0 || a.I >= len(g.options) {
			return ""
		}
		if contains(g.options[a.I].Owners, pid) {
			return "own"
		}
		if aud {
			g.crowd[pid] = a.I
		} else {
			g.picks[pid] = a.I
		}
	}
	return ""
}

func (g *Fibber) buildOptions(c *engine.Ctx) {
	byNorm := map[string]int{}
	real := g.qs[g.round].A
	g.options = []fibOption{{Text: real, Real: true}}
	byNorm[norm(real)] = 0
	for _, id := range c.Players() {
		t, ok := g.fakes[id]
		if !ok {
			continue
		}
		if i, dup := byNorm[norm(t)]; dup {
			g.options[i].Owners = append(g.options[i].Owners, id)
			continue
		}
		byNorm[norm(t)] = len(g.options)
		g.options = append(g.options, fibOption{Text: t, Owners: []string{id}})
	}
	for _, i := range c.Rng.Perm(len(g.pool)) {
		if len(g.options) >= 4 {
			break
		}
		if _, dup := byNorm[norm(g.pool[i])]; !dup {
			byNorm[norm(g.pool[i])] = len(g.options)
			g.options = append(g.options, fibOption{Text: g.pool[i]})
		}
	}
	c.Rng.Shuffle(len(g.options), func(i, j int) { g.options[i], g.options[j] = g.options[j], g.options[i] })
}

func (g *Fibber) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "write":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.fakes) {
			g.buildOptions(c)
			g.ph.Set("pick", c.Now, 30*time.Second)
		}
	case "pick":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.picks) {
			for pid, i := range g.picks {
				o := g.options[i]
				if o.Real {
					g.gain[pid] += 200
				}
				for _, owner := range o.Owners {
					g.gain[owner] += 100
				}
			}
			for _, i := range g.crowd {
				for _, owner := range g.options[i].Owners {
					g.gain[owner] += 50
				}
			}
			for id, v := range g.gain {
				c.Add(id, v)
			}
			g.ph.Set("reveal", c.Now, 9*time.Second)
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

func (g *Fibber) view(c *engine.Ctx) engine.View {
	v := map[string]any{"phase": g.ph.Name, "round": g.round + 1, "rounds": g.rounds, "ms": g.ph.Ms(c.Now), "q": g.qs[g.round].Q, "total": len(c.Active())}
	switch g.ph.Name {
	case "write":
		v["done"] = keys(g.fakes, c.Players())
	case "pick", "reveal":
		opts := make([]map[string]any, len(g.options))
		for i, o := range g.options {
			m := map[string]any{"text": o.Text}
			if g.ph.Name == "reveal" {
				m["real"], m["owners"] = o.Real, o.Owners
				var pickers []string
				for _, id := range c.Players() {
					if j, ok := g.picks[id]; ok && j == i {
						pickers = append(pickers, id)
					}
				}
				m["pickers"] = pickers
			}
			opts[i] = m
		}
		v["options"] = opts
		v["done"] = keys(g.picks, c.Players())
		if g.ph.Name == "reveal" {
			v["gain"] = g.gain
		}
	}
	if g.done {
		v["phase"] = "end"
	}
	return engine.View{Shared: v}
}

func (g *Fibber) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Fibber) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Fibber) Me(c *engine.Ctx, pid string) any {
	own := []int{}
	for i, o := range g.options {
		if contains(o.Owners, pid) {
			own = append(own, i)
		}
	}
	pick := -1
	if i, ok := g.picks[pid]; ok {
		pick = i
	} else if i, ok := g.crowd[pid]; ok {
		pick = i
	}
	return map[string]any{"fake": g.fakes[pid], "own": own, "pick": pick}
}

func (g *Fibber) Done() bool { return g.done }

func (g *Fibber) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	switch g.ph.Name {
	case "write":
		if _, ok := g.fakes[pid]; !ok && len(g.pool) > 1 && engine.BotDue(pid, g.ph, c.Now, 5000, 20000) {
			for _, i := range c.Rng.Perm(len(g.pool)) {
				if norm(g.pool[i]) != norm(g.qs[g.round].A) {
					return engine.Action{K: "submit", S: g.pool[i]}, true
				}
			}
		}
	case "pick":
		if _, ok := g.picks[pid]; !ok && engine.BotDue(pid, g.ph, c.Now, 3000, 12000) {
			for _, i := range c.Rng.Perm(len(g.options)) {
				if !contains(g.options[i].Owners, pid) {
					return engine.Action{K: "pick", I: i}, true
				}
			}
		}
	}
	return engine.Action{}, false
}
