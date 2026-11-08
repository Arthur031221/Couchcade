package games

import (
	"strings"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
)

// Mole: everyone but one player knows the secret word. Give a one word clue, then vote out the mole.
type Mole struct {
	ph      engine.Phase
	round   int
	rounds  int
	words   []content.Word
	pool    []content.Word
	moles   []string
	ids     []string // seated players when the game began
	clues   map[string]string
	votes   map[string]int
	guess   string
	guessed bool
	caught  bool
	gain    map[string]int
	done    bool
}

func (g *Mole) Info() engine.Info { return Catalog[5] }

func (g *Mole) Start(c *engine.Ctx) {
	g.ids = append([]string(nil), c.Players()...)
	g.rounds = c.Rounds(3)
	g.pool = c.Content.Mole(c.Lang, c.Family)
	for _, i := range pickN(c.Rng, max(len(g.pool), 1), g.rounds) {
		if len(g.pool) > 0 {
			g.words = append(g.words, g.pool[i])
		} else {
			g.words = append(g.words, content.Word{Category: "?", Word: "?"})
		}
	}
	order := shuffled(c.Rng, g.ids)
	for i := 0; i < g.rounds; i++ {
		g.moles = append(g.moles, order[i%len(order)])
	}
	g.round = -1
	g.nextRound(c)
}

func (g *Mole) nextRound(c *engine.Ctx) {
	g.round++
	g.clues, g.votes, g.gain, g.guess, g.guessed, g.caught = map[string]string{}, map[string]int{}, map[string]int{}, "", false, false
	g.ph.Set("clue", c.Now, 45*time.Second)
}

func (g *Mole) mole() string { return g.moles[g.round] }

func (g *Mole) Act(c *engine.Ctx, pid string, a engine.Action) string {
	if c.IsAudience(pid) || !contains(g.ids, pid) {
		return ""
	}
	switch {
	case a.K == "submit" && g.ph.Name == "clue":
		t := clean(a.S, 24)
		if f := strings.Fields(t); len(f) > 0 {
			t = f[0]
		}
		if norm(t) == "" {
			return "empty"
		}
		if pid != g.mole() && norm(t) == norm(g.words[g.round].Word) {
			return "secret_word"
		}
		g.clues[pid] = t
	case a.K == "vote" && g.ph.Name == "vote":
		if a.I < 0 || a.I >= len(g.ids) {
			return ""
		}
		if g.ids[a.I] == pid {
			return "own"
		}
		g.votes[pid] = a.I
	case a.K == "guess" && g.ph.Name == "guess" && pid == g.mole():
		g.guess, g.guessed = clean(a.S, 40), true
	}
	return ""
}

func (g *Mole) Tick(c *engine.Ctx) {
	switch g.ph.Name {
	case "clue":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.clues) {
			g.ph.Set("vote", c.Now, 45*time.Second)
		}
	case "vote":
		if g.ph.Over(c.Now) || allActed(c.Active(), g.votes) {
			g.resolve(c)
		}
	case "guess":
		if g.ph.Over(c.Now) || g.guessed {
			if norm(g.guess) == norm(g.words[g.round].Word) && g.guessed {
				g.gain[g.mole()] += 300
				c.Add(g.mole(), 300)
			}
			g.ph.Set("reveal", c.Now, 10*time.Second)
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

func (g *Mole) resolve(c *engine.Ctx) {
	tally := map[int]int{}
	for _, i := range g.votes {
		tally[i]++
	}
	top, topN, tie := -1, 0, false
	for i, n := range tally {
		switch {
		case n > topN:
			top, topN, tie = i, n, false
		case n == topN:
			tie = true
		}
	}
	m := g.mole()
	g.caught = !tie && top >= 0 && g.ids[top] == m
	for voter, i := range g.votes {
		if g.ids[i] == m {
			g.gain[voter] += 200
			c.Add(voter, 200)
		}
	}
	if g.caught {
		g.ph.Set("guess", c.Now, 25*time.Second)
		return
	}
	g.gain[m] += 400
	c.Add(m, 400)
	g.ph.Set("reveal", c.Now, 10*time.Second)
}

func (g *Mole) view(c *engine.Ctx) engine.View {
	w := g.words[g.round]
	v := map[string]any{"phase": g.ph.Name, "round": g.round + 1, "rounds": g.rounds, "ms": g.ph.Ms(c.Now),
		"category": w.Category, "ids": g.ids, "total": len(c.Active())}
	switch g.ph.Name {
	case "clue":
		v["done"] = keys(g.clues, g.ids)
	case "vote", "guess", "reveal":
		v["clues"] = g.clues
		v["done"] = keys(g.votes, g.ids)
	}
	if g.ph.Name == "guess" || g.ph.Name == "reveal" {
		v["caught"], v["mole"] = g.caught, g.mole()
	}
	if g.ph.Name == "reveal" {
		v["word"], v["guess"], v["gain"] = w.Word, g.guess, g.gain
		votes := map[string]string{}
		for voter, i := range g.votes {
			votes[voter] = g.ids[i]
		}
		v["votes"] = votes
	}
	if g.done {
		v["phase"] = "end"
	}
	return engine.View{Shared: v}
}

func (g *Mole) TVView(c *engine.Ctx) engine.View    { return g.view(c) }
func (g *Mole) PhoneView(c *engine.Ctx) engine.View { return g.view(c) }

func (g *Mole) Me(c *engine.Ctx, pid string) any {
	me := map[string]any{"mole": pid == g.mole(), "clue": g.clues[pid], "gain": g.gain[pid]}
	if pid != g.mole() {
		me["word"] = g.words[g.round].Word
	}
	if i, ok := g.votes[pid]; ok {
		me["vote"] = i
	} else {
		me["vote"] = -1
	}
	return me
}

func (g *Mole) Done() bool { return g.done }

func (g *Mole) Bot(c *engine.Ctx, pid string) (engine.Action, bool) {
	if c.IsAudience(pid) || !contains(g.ids, pid) {
		return engine.Action{}, false
	}
	switch g.ph.Name {
	case "clue":
		if _, ok := g.clues[pid]; !ok && engine.BotDue(pid, g.ph, c.Now, 4000, 18000) {
			return engine.Action{K: "submit", S: g.relatedWord(c, pid)}, true
		}
	case "vote":
		if _, ok := g.votes[pid]; !ok && engine.BotDue(pid, g.ph, c.Now, 4000, 16000) {
			for _, i := range c.Rng.Perm(len(g.ids)) {
				if g.ids[i] != pid {
					return engine.Action{K: "vote", I: i}, true
				}
			}
		}
	case "guess":
		if pid == g.mole() && !g.guessed && engine.BotDue(pid, g.ph, c.Now, 3000, 9000) {
			return engine.Action{K: "guess", S: g.relatedWord(c, pid)}, true
		}
	}
	return engine.Action{}, false
}

// relatedWord picks another word from the same category, which reads like a real clue.
func (g *Mole) relatedWord(c *engine.Ctx, pid string) string {
	w := g.words[g.round]
	var same []string
	for _, x := range g.pool {
		if x.Category == w.Category && norm(x.Word) != norm(w.Word) {
			same = append(same, x.Word)
		}
	}
	if len(same) == 0 {
		return w.Category
	}
	return same[c.Rng.Intn(len(same))]
}
