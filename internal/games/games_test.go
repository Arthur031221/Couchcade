package games_test

import (
	"testing"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
	"github.com/Arthur031221/Couchcade/internal/games"
)

func TestBotsPlayEveryGameToTheEnd(t *testing.T) {
	for _, info := range games.Catalog {
		t.Run(info.ID, func(t *testing.T) {
			h := newHarness(t, 0)
			h.host("bots", `,"n":4`)
			h.start(info.ID)
			if h.roomPhase() != "game" {
				t.Fatalf("game did not start, room phase %q", h.roomPhase())
			}
			if !h.until(20*time.Minute, func() bool { return h.roomPhase() == "final" }) {
				t.Fatalf("no final screen, stuck in %q", h.phase())
			}
			room, _ := h.tvState()
			total := 0
			for _, p := range room["players"].([]any) {
				total += int(p.(map[string]any)["score"].(float64))
			}
			if total == 0 {
				t.Fatal("nobody scored")
			}
			h.step(31 * time.Second)
			if h.roomPhase() != "lobby" {
				t.Fatalf("expected lobby after the final screen, got %q", h.roomPhase())
			}
		})
	}
}

func TestEchoScoresMatchingAnswers(t *testing.T) {
	h := newHarness(t, 3)
	h.start("echo")
	h.act(0, engine.Action{K: "submit", S: "Pizza"})
	h.act(1, engine.Action{K: "submit", S: " pizza! "})
	h.act(2, engine.Action{K: "submit", S: "dog"})
	h.waitPhase("reveal")
	if got := []int{h.score(0), h.score(1), h.score(2)}; got[0] != 100 || got[1] != 100 || got[2] != 0 {
		t.Fatalf("scores %v", got)
	}
}

func fibberAnswer(t *testing.T, q string) string {
	set, _ := content.Builtin()
	for _, it := range set.Fibber("en", true) {
		if it.Q == q {
			return it.A
		}
	}
	t.Fatalf("question %q not in pack", q)
	return ""
}

func TestFibberRejectsRealAnswerAndScores(t *testing.T) {
	h := newHarness(t, 3)
	h.start("fibber")
	real := fibberAnswer(t, h.tvg()["q"].(string))
	h.act(0, engine.Action{K: "submit", S: real})
	if h.errCode(0) != "real_answer" {
		t.Fatalf("expected real_answer, got %q", h.errCode(0))
	}
	h.act(0, engine.Action{K: "submit", S: "Alpha"})
	h.act(1, engine.Action{K: "submit", S: "Beta"})
	h.act(2, engine.Action{K: "submit", S: "Gamma"})
	h.waitPhase("pick")
	idx := map[string]int{}
	for i, o := range h.tvg()["options"].([]any) {
		idx[o.(map[string]any)["text"].(string)] = i
	}
	h.act(0, engine.Action{K: "pick", I: idx[real]})
	h.act(1, engine.Action{K: "pick", I: idx["Alpha"]})
	h.act(2, engine.Action{K: "pick", I: idx[real]})
	h.act(2, engine.Action{K: "pick", I: idx["Gamma"]}) // picking your own fake must be refused
	if h.errCode(2) != "own" {
		t.Fatalf("expected own, got %q", h.errCode(2))
	}
	h.waitPhase("reveal")
	if got := []int{h.score(0), h.score(1), h.score(2)}; got[0] != 300 || got[1] != 0 || got[2] != 200 {
		t.Fatalf("scores %v", got)
	}
}

func TestTriviaSpeedBonusAndTeams(t *testing.T) {
	h := newHarness(t, 4)
	h.start("trivia")
	set, _ := content.Builtin()
	var answer string
	q := h.tvg()["q"].(string)
	for _, it := range set.Trivia("en", true) {
		if it.Q == q {
			answer = it.Options[it.Answer]
		}
	}
	correct, wrong := -1, -1
	for i, o := range h.tvg()["options"].([]any) {
		if o.(string) == answer {
			correct = i
		} else {
			wrong = i
		}
	}
	h.act(0, engine.Action{K: "answer", I: correct})
	h.act(1, engine.Action{K: "answer", I: wrong})
	h.act(2, engine.Action{K: "answer", I: wrong})
	h.act(3, engine.Action{K: "answer", I: wrong})
	h.waitPhase("reveal")
	if s := h.score(0); s < 190 || s > 200 {
		t.Fatalf("fast correct answer scored %d", s)
	}
	if h.score(1) != 0 {
		t.Fatal("wrong answer scored")
	}
	teams := h.tvg()["teams"].(map[string]any)
	a, b := 0, 0
	for _, v := range teams {
		if v.(float64) == 0 {
			a++
		} else {
			b++
		}
	}
	if a != 2 || b != 2 {
		t.Fatalf("teams not balanced: %d vs %d", a, b)
	}
}

func TestZapEarlyTapCostsAndFastestWins(t *testing.T) {
	h := newHarness(t, 3)
	h.start("zap")
	h.act(0, engine.Action{K: "tap"})
	h.waitPhase("go")
	h.act(2, engine.Action{K: "tap"})
	h.step(200 * time.Millisecond)
	h.act(1, engine.Action{K: "tap"})
	h.waitPhase("reveal")
	if got := []int{h.score(0), h.score(1), h.score(2)}; got[0] != -100 || got[1] != 200 || got[2] != 300 {
		t.Fatalf("scores %v", got)
	}
}

func moleIndex(h *harness, n int) int {
	for i := 0; i < n; i++ {
		if h.gme(i)["mole"] == true {
			return i
		}
	}
	return -1
}

func TestMoleCaughtAndEscaped(t *testing.T) {
	h := newHarness(t, 4)
	h.start("mole")
	m := moleIndex(h, 4)
	if m < 0 {
		t.Fatal("no mole")
	}
	if _, has := h.gme(m)["word"]; has {
		t.Fatal("the mole must not see the word")
	}
	word := ""
	for i := 0; i < 4; i++ {
		if i != m {
			word = h.gme(i)["word"].(string)
		}
	}
	h.act((m+1)%4, engine.Action{K: "submit", S: word})
	if h.errCode((m+1)%4) != "secret_word" {
		t.Fatal("an agent must not be able to type the secret word")
	}
	for i := 0; i < 4; i++ {
		h.act(i, engine.Action{K: "submit", S: "clue" + string(rune('a'+i))})
	}
	h.waitPhase("vote")
	for i := 0; i < 4; i++ {
		target := m
		if i == m {
			target = (m + 1) % 4
		}
		h.act(i, engine.Action{K: "vote", I: target})
	}
	h.waitPhase("guess")
	h.act(m, engine.Action{K: "guess", S: word})
	h.waitPhase("reveal")
	if h.score(m) != 300 {
		t.Fatalf("a mole that guesses the word after being caught gets 300, got %d", h.score(m))
	}
	if h.score((m+1)%4) != 200 {
		t.Fatalf("an agent who found the mole gets 200, got %d", h.score((m+1)%4))
	}
	// Next round: nobody agrees, the mole walks away with 400.
	h.waitPhase("clue")
	m2 := moleIndex(h, 4)
	before := h.score(m2)
	for i := 0; i < 4; i++ {
		h.act(i, engine.Action{K: "submit", S: "x" + string(rune('a'+i))})
	}
	h.waitPhase("vote")
	for i := 0; i < 4; i++ {
		h.act(i, engine.Action{K: "vote", I: (i + 1) % 4})
	}
	h.waitPhase("reveal")
	if h.score(m2)-before != 400 {
		t.Fatalf("an escaped mole gets 400, got %d", h.score(m2)-before)
	}
}

func TestSquiggleVotingAndAudience(t *testing.T) {
	h := newHarness(t, 8)
	aud := h.join("Watcher")
	if h.me(aud)["aud"] != true {
		t.Fatal("the ninth phone should join as audience")
	}
	h.start("squiggle")
	dot := [][]int{{3, 8, 100, 100, 400, 400}}
	for i := 0; i < 3; i++ {
		h.act(i, engine.Action{K: "submit", P: dot})
	}
	h.act(3, engine.Action{K: "submit", P: [][]int{{1}}}) // malformed: refused
	if h.errCode(3) != "empty" {
		t.Fatal("a malformed drawing must be refused")
	}
	h.step(76 * time.Second)
	h.waitPhase("vote")
	own := make([]int, 3)
	for i := range own {
		own[i] = int(h.gme(i)["own"].(float64))
	}
	h.act(0, engine.Action{K: "vote", I: own[1]})
	h.act(2, engine.Action{K: "vote", I: own[1]})
	h.act(1, engine.Action{K: "vote", I: own[1]})
	if h.errCode(1) != "own" {
		t.Fatal("voting for your own drawing must be refused")
	}
	h.act(1, engine.Action{K: "vote", I: own[0]})
	h.act(aud, engine.Action{K: "vote", I: own[2]})
	h.waitPhase("reveal")
	// P2: two votes (200) + round winner bonus (100). P1: one vote. P3: one crowd vote.
	if got := []int{h.score(0), h.score(1), h.score(2)}; got[0] != 100 || got[1] != 300 || got[2] != 50 {
		t.Fatalf("scores %v", got)
	}
}
