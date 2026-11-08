package web

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func load(t *testing.T, name string) map[string]string {
	t.Helper()
	b, err := fs.ReadFile(FS, "locales/"+name)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return m
}

// Every locale must define exactly the keys English defines.
func TestLocalesMatchEnglishKeys(t *testing.T) {
	en := load(t, "en.json")
	files, err := fs.Glob(FS, "locales/*.json")
	if err != nil || len(files) != 9 {
		t.Fatalf("want 9 locale files, got %d (%v)", len(files), err)
	}
	for _, f := range files {
		name := f[len("locales/"):]
		m := load(t, name)
		for k := range en {
			if m[k] == "" {
				t.Errorf("%s misses key %s", name, k)
			}
		}
		for k := range m {
			if _, ok := en[k]; !ok {
				t.Errorf("%s has unknown key %s", name, k)
			}
		}
	}
}
