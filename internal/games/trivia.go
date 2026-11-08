package games

import (
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
)

const triviaTime = 20 * time.Second

// Trivia: two teams answer the same multiple choice questions, faster is worth more.
type Trivia struct {
	ph      engine.Phase
	idx     int
	count   int
	qs      []content.Quiz
	team    map[string]int
	answers map[string]int
	crowd   map[string]int
	gain    map[string]int
	totals  [2]int
	done    bool
}

func (g *Trivia) Info() engine.Info { return Catalog[2] }

func (g *Trivia) Start(c *engine.Ctx) {
	g.count = c.Rounds(8)
	items := c.Content.Trivia(c.Lang, c.Family)
	for _, i := range pickN(c.Rng, max(len(items), 1), g.count) {
		if len(items) > 0 {
			g.qs = append(g.qs, shuffleOptions(c, items[i]))
		}
	}
	g.count = len(g.qs)
	g.team = map[string]int{}
	for i, id := range shuffled(c.Rng, c.Players()) {
		g.team[id] = i % 2
	}
	g.idx = -1
	g.next(c)
}

func (g *Trivia) next(c *engine.Ctx) {
	g.idx++
	g.answers, g.crowd, g.gain = map[string]int{}, map[string]int{}, map[string]int{}
	g.ph.Set("ask", c.Now, triviaTime)
}

func (g *Trivia) Act(c *engine.Ctx, pid string, a engine.Action) string {
	if a.K != "answer" || g.ph.Name != "ask" || a.I < 0 || a.I > 3 {
		return ""
	}
	if c.IsAudience(pid) {
		g.crowd[pid] = a.I
		return ""
	}
	if _, again := g.answers[pid]; !again {
		g.answers[pid] = a.I
		if a.I == g.qs[g.idx].Answer {
			left := float64(g.ph.Until.Sub(c.Now)) / float64(triviaTime)
			g.gain[pid] = 100 + int(100*max(left, 0))
		}
	}
	return ""
}

func (g *Trivia) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "ask":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.answers) {
			for id, v := range g.gain {
				c.Add(id, v)
				g.totals[g.team[id]] += v
			}
			g.ph.Set("reveal", c.Now, 6*time.Second)
		}
	case "reveal":
		if g.ph.Over(c.Now) {
			if g.idx+1 >= g.count {
				g.done = true
				return
			}
			g.next(c)
		}
	}
}

func (g *Trivia) view(c *engine.Ctx) engine.View {
	q := g.qs[g.idx]
	v := map[string]any{"phase": g.ph.Name, "round": g.idx + 1, "rounds": g.count, "ms": g.ph.Ms(c.Now),
		"q": q.Q, "options": q.Options, "teams": g.team, "totals": g.totals, "done": keys(g.answers, c.Players()),
		"total": len(c.Active()), "crowdN": len(g.crowd)}
	if g.ph.Name == "reveal" || g.done {
		counts, crowd := [4]int{}, [4]int{}
		for _, i := range g.answers {
			counts[i]++
		}
		for _, i := range g.crowd {
			crowd[i]++
		}
		v["answer"], v["counts"], v["crowd"], v["gain"] = q.Answer, counts, crowd, g.gain
	}
	if g.done {
		v["phase"] = "end"
	}
	return engine.View{Shared: v}
}

func (g *Trivia) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Trivia) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Trivia) Me(c *engine.Ctx, pid string) any {
	pick := -1
	if i, ok := g.answers[pid]; ok {
		pick = i
	}
	if i, ok := g.crowd[pid]; ok {
		pick = i
	}
	return map[string]any{"pick": pick, "team": g.team[pid], "gain": g.gain[pid]}
}

func (g *Trivia) Done() bool { return g.done }

func (g *Trivia) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	if g.ph.Name != "ask" || !engine.BotDue(pid, g.ph, c.Now, 2500, 14000) {
		return engine.Action{}, false
	}
	if _, ok := g.answers[pid]; ok {
		return engine.Action{}, false
	}
	if _, ok := g.crowd[pid]; ok {
		return engine.Action{}, false
	}
	i := g.qs[g.idx].Answer
	if c.Rng.Intn(100) >= 55 {
		i = c.Rng.Intn(4)
	}
	return engine.Action{K: "answer", I: i}, true
}

// shuffleOptions randomizes the answer position so packs need no balancing.
func shuffleOptions(c *engine.Ctx, q content.Quiz) content.Quiz {
	out := q
	out.Options = make([]string, len(q.Options))
	for to, from := range c.Rng.Perm(len(q.Options)) {
		out.Options[to] = q.Options[from]
		if from == q.Answer {
			out.Answer = to
		}
	}
	return out
}
