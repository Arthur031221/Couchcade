package games

import (
	"sort"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
)

// Echo: everyone answers the same prompt, and you score for every player who wrote the same thing.
type Echo struct {
	ph      engine.Phase
	round   int
	rounds  int
	items   []content.Text
	answers map[string]string
	groups  []echoGroup
	gain    map[string]int
	done    bool
}

type echoGroup struct {
	Text string   `json:"text"`
	IDs  []string `json:"ids"`
}

func (g *Echo) Info() engine.Info { return Catalog[4] }

func (g *Echo) Start(c *engine.Ctx) {
	g.rounds = c.Rounds(5)
	items := c.Content.Echo(c.Lang, c.Family)
	for _, i := range pickN(c.Rng, max(len(items), 1), g.rounds) {
		if len(items) > 0 {
			g.items = append(g.items, items[i])
		} else {
			g.items = append(g.items, content.Text{Text: "?"})
		}
	}
	g.round = -1
	g.nextRound(c)
}

func (g *Echo) nextRound(c *engine.Ctx) {
	g.round++
	g.answers, g.gain, g.groups = map[string]string{}, map[string]int{}, nil
	g.ph.Set("answer", c.Now, 30*time.Second)
}

func (g *Echo) Act(c *engine.Ctx, pid string, a engine.Action) string {
	if a.K != "submit" || g.ph.Name != "answer" || c.IsAudience(pid) {
		return ""
	}
	t := clean(a.S, 30)
	if norm(t) == "" {
		return "empty"
	}
	g.answers[pid] = t
	return ""
}

func (g *Echo) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "answer":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.answers) {
			g.group(c)
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

func (g *Echo) group(c *engine.Ctx) {
	idx := map[string]int{}
	for _, id := range c.Players() {
		t, ok := g.answers[id]
		if !ok {
			continue
		}
		k := norm(t)
		i, seen := idx[k]
		if !seen {
			i = len(g.groups)
			idx[k] = i
			g.groups = append(g.groups, echoGroup{Text: t})
		}
		g.groups[i].IDs = append(g.groups[i].IDs, id)
	}
	sort.SliceStable(g.groups, func(i, j int) bool { return len(g.groups[i].IDs) > len(g.groups[j].IDs) })
	for _, gr := range g.groups {
		for _, id := range gr.IDs {
			g.gain[id] = (len(gr.IDs) - 1) * 100
			c.Add(id, g.gain[id])
		}
	}
}

func (g *Echo) view(c *engine.Ctx) engine.View {
	v := map[string]any{"phase": g.ph.Name, "round": g.round + 1, "rounds": g.rounds, "ms": g.ph.Ms(c.Now),
		"prompt": g.items[g.round].Text, "done": keys(g.answers, c.Players()), "total": len(c.Active())}
	if g.ph.Name == "reveal" {
		v["groups"], v["gain"] = g.groups, g.gain
	}
	if g.done {
		v["phase"] = "end"
	}
	return engine.View{Shared: v}
}

func (g *Echo) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Echo) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Echo) Me(c *engine.Ctx, pid string) any {
	return map[string]any{"text": g.answers[pid], "gain": g.gain[pid]}
}

func (g *Echo) Done() bool { return g.done }

func (g *Echo) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	if g.ph.Name != "answer" || !engine.BotDue(pid, g.ph, c.Now, 3000, 16000) {
		return engine.Action{}, false
	}
	if _, ok := g.answers[pid]; ok {
		return engine.Action{}, false
	}
	hints := g.items[g.round].Hints
	if len(hints) == 0 {
		hints = []string{"pizza", "dog", "blue", "sun"}
	}
	i := 0
	if c.Rng.Intn(100) >= 45 {
		i = c.Rng.Intn(len(hints))
	}
	return engine.Action{K: "submit", S: hints[i]}, true
}
