// Package games holds the six game state machines.
package games

import (
	"math"
	"math/rand"
	"strings"
	"unicode"

	"github.com/Arthur031221/Couchcade/internal/engine"
)

// norm folds a typed answer so "Ice  cream!" and "icecream" compare equal.
func norm(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func clean(s string, max int) string {
	var b strings.Builder
	for _, r := range strings.Join(strings.Fields(s), " ") {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	rs := []rune(b.String())
	if len(rs) > max {
		rs = rs[:max]
	}
	return string(rs)
}

func allActed[V any](ids []string, m map[string]V) bool {
	if len(ids) == 0 {
		return false // nobody to wait for means wait for the clock, not skip the phase
	}
	for _, id := range ids {
		if _, ok := m[id]; !ok {
			return false
		}
	}
	return true
}

func keys[V any](m map[string]V, order []string) []string {
	out := []string{}
	for _, id := range order {
		if _, ok := m[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func shuffled(r *rand.Rand, ids []string) []string {
	out := append([]string(nil), ids...)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// pickN returns n distinct indexes below size, repeating only when size < n.
func pickN(r *rand.Rand, size, n int) []int {
	out := make([]int, 0, n)
	for len(out) < n {
		out = append(out, r.Perm(size)...)
	}
	return out[:n]
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// Catalog lists the games in lobby order.
var Catalog = []engine.Info{
	{ID: "squiggle", Min: 2, Rounds: 3},
	{ID: "fibber", Min: 2, Rounds: 3},
	{ID: "trivia", Min: 2, Rounds: 8},
	{ID: "zap", Min: 2, Rounds: 5},
	{ID: "echo", Min: 2, Rounds: 5},
	{ID: "mole", Min: 3, Rounds: 3},
}

// New returns a fresh game by id, or nil.
func New(id string) engine.Game {
	switch id {
	case "squiggle":
		return &Squiggle{}
	case "fibber":
		return &Fibber{}
	case "trivia":
		return &Trivia{}
	case "zap":
		return &Zap{}
	case "echo":
		return &Echo{}
	case "mole":
		return &Mole{}
	}
	return nil
}

func sinf(x float64) float64 { return math.Sin(x) }
func cosf(x float64) float64 { return math.Cos(x) }
