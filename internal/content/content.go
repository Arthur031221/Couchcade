// Package content loads and validates the prompt packs the games draw from.
package content

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed packs/*.json
var builtin embed.FS

// Rating marks whether an item is fine for every audience.
const (
	RatingFamily = "family"
	RatingMature = "mature"
)

// Languages lists the interface languages the project ships.
var Languages = []string{"en", "zh-TW", "zh-CN", "ja", "ko", "es", "fr", "de", "pt-BR"}

type Text struct {
	Text   string `json:"text"`
	Rating string `json:"rating,omitempty"`
	// Hints are typical answers; simulated phones use them.
	Hints []string `json:"hints,omitempty"`
}

type QA struct {
	Q      string `json:"q"`
	A      string `json:"a"`
	Rating string `json:"rating,omitempty"`
}

type Quiz struct {
	Q       string   `json:"q"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
	Rating  string   `json:"rating,omitempty"`
}

type Word struct {
	Category string `json:"category"`
	Word     string `json:"word"`
	Rating   string `json:"rating,omitempty"`
}

type Items struct {
	Squiggle []Text `json:"squiggle,omitempty"`
	Fibber   []QA   `json:"fibber,omitempty"`
	Trivia   []Quiz `json:"trivia,omitempty"`
	Echo     []Text `json:"echo,omitempty"`
	Mole     []Word `json:"mole,omitempty"`
}

// Pack is one JSON file: prompts for one language.
type Pack struct {
	ID     string `json:"id"`
	Lang   string `json:"lang"`
	Name   string `json:"name,omitempty"`
	Author string `json:"author,omitempty"`
	Items  Items  `json:"items"`
}

// Set is every loaded pack, merged per language.
type Set struct {
	byLang map[string]*Items
}

func NewSet() *Set { return &Set{byLang: map[string]*Items{}} }

// Add merges a validated pack into the set.
func (s *Set) Add(p *Pack) {
	it := s.byLang[p.Lang]
	if it == nil {
		it = &Items{}
		s.byLang[p.Lang] = it
	}
	it.Squiggle = append(it.Squiggle, p.Items.Squiggle...)
	it.Fibber = append(it.Fibber, p.Items.Fibber...)
	it.Trivia = append(it.Trivia, p.Items.Trivia...)
	it.Echo = append(it.Echo, p.Items.Echo...)
	it.Mole = append(it.Mole, p.Items.Mole...)
}

// Builtin returns a set holding the packs compiled into the binary.
func Builtin() (*Set, error) {
	s := NewSet()
	entries, err := builtin.ReadDir("packs")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		data, err := builtin.ReadFile("packs/" + e.Name())
		if err != nil {
			return nil, err
		}
		p, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		s.Add(p)
	}
	return s, nil
}

// LoadDir adds every *.json pack in dir to the set.
func (s *Set) LoadDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		p, err := Parse(data)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		s.Add(p)
	}
	return nil
}

// Parse decodes and validates one pack.
func Parse(data []byte) (*Pack, error) {
	var p Pack
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	if errs := Validate(&p); len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return &p, nil
}

func ok(r string) bool { return r == "" || r == RatingFamily || r == RatingMature }

func knownLang(l string) bool {
	for _, k := range Languages {
		if k == l {
			return true
		}
	}
	return false
}

// Validate returns every problem found in the pack, empty when it is usable.
func Validate(p *Pack) []string {
	var errs []string
	bad := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	if p.ID == "" {
		bad("id is required")
	}
	if !knownLang(p.Lang) {
		bad("lang %q is not one of %s", p.Lang, strings.Join(Languages, ", "))
	}
	seen := map[string]bool{}
	dup := func(game, key string) bool {
		k := game + "\x00" + strings.ToLower(strings.TrimSpace(key))
		if seen[k] {
			return true
		}
		seen[k] = true
		return false
	}
	short := func(s string, max int) bool {
		n := len([]rune(strings.TrimSpace(s)))
		return n > 0 && n <= max
	}
	for i, t := range p.Items.Squiggle {
		if !short(t.Text, 60) || !ok(t.Rating) || dup("squiggle", t.Text) {
			bad("squiggle[%d] %q: empty, over 60 characters, duplicate or bad rating", i, t.Text)
		}
	}
	for i, t := range p.Items.Echo {
		if !short(t.Text, 80) || !ok(t.Rating) || dup("echo", t.Text) {
			bad("echo[%d] %q: empty, over 80 characters, duplicate or bad rating", i, t.Text)
		}
	}
	for i, q := range p.Items.Fibber {
		if !short(q.Q, 200) || !short(q.A, 40) || !ok(q.Rating) || dup("fibber", q.Q) {
			bad("fibber[%d] %q: needs a question up to 200 and an answer up to 40 characters", i, q.Q)
		}
	}
	for i, q := range p.Items.Trivia {
		opts := map[string]bool{}
		for _, o := range q.Options {
			opts[strings.ToLower(strings.TrimSpace(o))] = true
			if !short(o, 60) {
				bad("trivia[%d] has an empty or long option", i)
			}
		}
		if !short(q.Q, 200) || len(q.Options) != 4 || len(opts) != 4 || q.Answer < 0 || q.Answer > 3 || !ok(q.Rating) || dup("trivia", q.Q) {
			bad("trivia[%d] %q: needs a question, four distinct options, answer 0 to 3", i, q.Q)
		}
	}
	for i, w := range p.Items.Mole {
		if !short(w.Category, 40) || !short(w.Word, 40) || !ok(w.Rating) || dup("mole", w.Word) {
			bad("mole[%d] %q: needs a category and a word", i, w.Word)
		}
	}
	return errs
}

func familyOnly[T any](in []T, family bool, rating func(T) string) []T {
	if !family {
		return in
	}
	out := make([]T, 0, len(in))
	for _, x := range in {
		if rating(x) != RatingMature {
			out = append(out, x)
		}
	}
	return out
}

// pick returns the items for lang, falling back to English when a language has none.
func pick[T any](s *Set, lang string, family bool, get func(*Items) []T, rating func(T) string) []T {
	for _, l := range []string{lang, "en"} {
		if it := s.byLang[l]; it != nil {
			if out := familyOnly(get(it), family, rating); len(out) > 0 {
				return out
			}
		}
	}
	return nil
}

func (s *Set) Squiggle(lang string, family bool) []Text {
	return pick(s, lang, family, func(i *Items) []Text { return i.Squiggle }, func(t Text) string { return t.Rating })
}
func (s *Set) Echo(lang string, family bool) []Text {
	return pick(s, lang, family, func(i *Items) []Text { return i.Echo }, func(t Text) string { return t.Rating })
}
func (s *Set) Fibber(lang string, family bool) []QA {
	return pick(s, lang, family, func(i *Items) []QA { return i.Fibber }, func(t QA) string { return t.Rating })
}
func (s *Set) Trivia(lang string, family bool) []Quiz {
	return pick(s, lang, family, func(i *Items) []Quiz { return i.Trivia }, func(t Quiz) string { return t.Rating })
}
func (s *Set) Mole(lang string, family bool) []Word {
	return pick(s, lang, family, func(i *Items) []Word { return i.Mole }, func(t Word) string { return t.Rating })
}

// Counts reports how many items each language has per game, for docs and the CLI.
func (s *Set) Counts() map[string]map[string]int {
	out := map[string]map[string]int{}
	for l, it := range s.byLang {
		out[l] = map[string]int{"squiggle": len(it.Squiggle), "fibber": len(it.Fibber), "trivia": len(it.Trivia), "echo": len(it.Echo), "mole": len(it.Mole)}
	}
	return out
}
