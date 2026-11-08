package content

import (
	"strings"
	"testing"
)

func TestBuiltinPacksAreValidAndComplete(t *testing.T) {
	s, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	counts := s.Counts()
	for _, l := range Languages {
		c, ok := counts[l]
		if !ok {
			t.Fatalf("no pack for %s", l)
		}
		want := map[string]int{"squiggle": 20, "fibber": 10, "trivia": 12, "echo": 12, "mole": 12}
		for game, n := range want {
			if c[game] < n {
				t.Errorf("%s has %d %s items, want at least %d", l, c[game], game, n)
			}
		}
	}
}

func TestParseRejectsBadPacks(t *testing.T) {
	cases := map[string]string{
		"unknown lang":  `{"id":"x","lang":"xx","items":{}}`,
		"unknown field": `{"id":"x","lang":"en","itemz":{}}`,
		"answer range":  `{"id":"x","lang":"en","items":{"trivia":[{"q":"q","options":["a","b","c","d"],"answer":4}]}}`,
		"dup options":   `{"id":"x","lang":"en","items":{"trivia":[{"q":"q","options":["a","a","c","d"],"answer":0}]}}`,
		"long prompt":   `{"id":"x","lang":"en","items":{"squiggle":[{"text":"` + strings.Repeat("a", 61) + `"}]}}`,
		"bad rating":    `{"id":"x","lang":"en","items":{"echo":[{"text":"t","rating":"spicy"}]}}`,
		"duplicate":     `{"id":"x","lang":"en","items":{"echo":[{"text":"t"},{"text":"T"}]}}`,
	}
	for name, js := range cases {
		if _, err := Parse([]byte(js)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestFamilyFilterDropsMatureItems(t *testing.T) {
	p, err := Parse([]byte(`{"id":"x","lang":"en","items":{"echo":[{"text":"a"},{"text":"b","rating":"mature"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	s := NewSet()
	s.Add(p)
	if n := len(s.Echo("en", true)); n != 1 {
		t.Fatalf("family mode kept %d items", n)
	}
	if n := len(s.Echo("en", false)); n != 2 {
		t.Fatalf("mature mode kept %d items", n)
	}
}

func TestMissingLanguageFallsBackToEnglish(t *testing.T) {
	p, _ := Parse([]byte(`{"id":"x","lang":"en","items":{"echo":[{"text":"a"}]}}`))
	s := NewSet()
	s.Add(p)
	if len(s.Echo("ja", true)) != 1 {
		t.Fatal("no fallback")
	}
}
