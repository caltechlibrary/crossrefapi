package crossrefapi

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// writeOverride writes body to a new override file and returns its path.
func writeOverride(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "symbols-override.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// TestDefaultSymbols_Loads checks the embedded table has every section and the
// entries the real CrossRef fixtures need (tokens and Greek letters seen in
// testdata/crossref-*.json, the invisible operators, common entities).
func TestDefaultSymbols_Loads(t *testing.T) {
	s, err := DefaultSymbols()
	if err != nil {
		t.Fatalf("DefaultSymbols() error: %v", err)
	}
	if s == nil {
		t.Fatal("DefaultSymbols() returned nil")
	}
	for name, n := range map[string]int{
		"tokens": len(s.Tokens), "entities": len(s.Entities),
		"superscripts": len(s.Superscripts), "subscripts": len(s.Subscripts),
		"tex": len(s.TeX), "dropped_headings": len(s.DroppedHeadings),
		"accents": len(s.Accents), "mathvariants": len(s.MathVariants),
	} {
		if n == 0 {
			t.Errorf("section %s is empty", name)
		}
	}
	want := []struct {
		section string
		m       map[string]string
		key     string
		val     string
	}{
		{"tokens", s.Tokens, "⁢", ""}, // invisible times is removed
		{"tokens", s.Tokens, "⁡", ""}, // function application
		{"entities", s.Entities, "nbsp", " "},
		{"entities", s.Entities, "thinsp", " "},
		{"entities", s.Entities, "minus", "−"},
		{"superscripts", s.Superscripts, "2", "²"},
		{"superscripts", s.Superscripts, "−", "⁻"}, // U+2212 minus
		{"subscripts", s.Subscripts, "2", "₂"},
		{"subscripts", s.Subscripts, "i", "ᵢ"},
		{"tex", s.TeX, "τ", `\tau`},
		{"tex", s.TeX, "ν", `\nu`},
		{"tex", s.TeX, "Λ", `\Lambda`},
		{"tex", s.TeX, "×", `\times`},
		{"tex", s.TeX, "∼", `\sim`},
		{"tex", s.TeX, "→", `\to`},
		{"tex", s.TeX, "±", `\pm`},
		{"accents", s.Accents, "¯", `\overline`},
		{"accents", s.Accents, "^", `\hat`},
		{"mathvariants", s.MathVariants, "double-struck", `\mathbb`},
		{"mathvariants", s.MathVariants, "normal", `\mathrm`},
	}
	for _, w := range want {
		got, ok := w.m[w.key]
		if !ok {
			t.Errorf("%s: missing key %q", w.section, w.key)
		} else if got != w.val {
			t.Errorf("%s[%q] = %q, want %q", w.section, w.key, got, w.val)
		}
	}
	if !contains(s.DroppedHeadings, "Abstract") {
		t.Errorf("dropped_headings = %v, want it to hold \"Abstract\"", s.DroppedHeadings)
	}
}

// TestDefaultSymbols_Invariants checks the shape rules independently of the
// loader's own validation: super/subscripts map one character to one character,
// and no key or TeX value is empty.
func TestDefaultSymbols_Invariants(t *testing.T) {
	s, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	for name, m := range map[string]map[string]string{"superscripts": s.Superscripts, "subscripts": s.Subscripts} {
		for k, v := range m {
			if utf8.RuneCountInString(k) != 1 || utf8.RuneCountInString(v) != 1 {
				t.Errorf("%s[%q] = %q: want one character to one character", name, k, v)
			}
		}
	}
	for name, m := range map[string]map[string]string{"tokens": s.Tokens, "entities": s.Entities, "tex": s.TeX} {
		for k := range m {
			if k == "" {
				t.Errorf("%s has an empty key", name)
			}
		}
	}
	for k, v := range s.TeX {
		if v == "" {
			t.Errorf("tex[%q] is empty", k)
		}
	}
	for name, m := range map[string]map[string]string{"accents": s.Accents, "mathvariants": s.MathVariants} {
		for k, v := range m {
			if k == "" || !strings.HasPrefix(v, `\`) {
				t.Errorf("%s[%q] = %q: want a non-empty key and a TeX command", name, k, v)
			}
		}
	}
	for k, v := range s.Entities {
		if strings.ContainsAny(k, "&;") {
			t.Errorf("entities key %q holds & or ;; keys are bare names", k)
		}
		if v == "" {
			t.Errorf("entities[%q] is empty", k)
		}
	}
}

// TestDefaultSymbols_FreshCopy checks each call returns an independent value,
// so a caller changing its table cannot change the next caller's.
func TestDefaultSymbols_FreshCopy(t *testing.T) {
	a, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	a.TeX["τ"] = "CHANGED"
	a.DroppedHeadings[0] = "CHANGED"
	b, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	if b.TeX["τ"] != `\tau` {
		t.Errorf("second call saw the first call's change: %q", b.TeX["τ"])
	}
	if b.DroppedHeadings[0] == "CHANGED" {
		t.Error("second call saw the first call's change to dropped_headings")
	}
}

// TestLoadSymbols_NoOverride checks an empty path means "no override": the
// result is the default table and there is no error.
func TestLoadSymbols_NoOverride(t *testing.T) {
	got, err := LoadSymbols("")
	if err != nil {
		t.Fatalf("LoadSymbols(\"\") error: %v", err)
	}
	want, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Error("LoadSymbols(\"\") differs from DefaultSymbols()")
	}
}

// TestLoadSymbols_OverrideReplacesAndAdds checks an override entry replaces a
// default of the same key, a new key is added, and everything else is kept.
func TestLoadSymbols_OverrideReplacesAndAdds(t *testing.T) {
	p := writeOverride(t, `{
  "tex": {"τ": "\\mathrm{tau}", "ℵ": "\\aleph"},
  "tokens": {"⁢": "*", "x": ""},
  "subscripts": {"⊙": "⊙"},
  "accents": {"^": "\\widehat", "←": "\\overleftarrow"},
  "mathvariants": {"bold": "\\boldsymbol", "frak": "\\mathfrak"}
}`)
	def, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadSymbols(p)
	if err != nil {
		t.Fatalf("LoadSymbols error: %v", err)
	}
	if got.TeX["τ"] != `\mathrm{tau}` {
		t.Errorf("replace: tex[τ] = %q", got.TeX["τ"])
	}
	if got.TeX["ℵ"] != `\aleph` {
		t.Errorf("add: tex[ℵ] = %q", got.TeX["ℵ"])
	}
	if got.TeX["ν"] != def.TeX["ν"] {
		t.Errorf("untouched default tex[ν] = %q, want %q", got.TeX["ν"], def.TeX["ν"])
	}
	if len(got.TeX) != len(def.TeX)+1 {
		t.Errorf("tex has %d entries, want %d", len(got.TeX), len(def.TeX)+1)
	}
	if got.Tokens["⁢"] != "*" {
		t.Errorf("replacing a token with a non-empty value: %q", got.Tokens["⁢"])
	}
	if v, ok := got.Tokens["x"]; !ok || v != "" {
		t.Errorf("a token mapped to the empty string must be kept: %q, %v", v, ok)
	}
	if got.Subscripts["⊙"] != "⊙" {
		t.Errorf("added subscript = %q", got.Subscripts["⊙"])
	}
	if got.Accents["^"] != `\widehat` || got.Accents["¯"] != def.Accents["¯"] || len(got.Accents) != len(def.Accents)+1 {
		t.Errorf("accents override: %v", got.Accents)
	}
	if got.MathVariants["bold"] != `\boldsymbol` || got.MathVariants["frak"] != `\mathfrak` || got.MathVariants["double-struck"] != def.MathVariants["double-struck"] {
		t.Errorf("mathvariants override: %v", got.MathVariants)
	}
	if got.Superscripts["2"] != def.Superscripts["2"] || len(got.Entities) != len(def.Entities) {
		t.Error("sections the override did not name were changed")
	}
}

// TestLoadSymbols_DroppedHeadings checks the override's list replaces the
// default list, an empty list empties it, and an absent key keeps the default.
func TestLoadSymbols_DroppedHeadings(t *testing.T) {
	def, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, body string
		want       []string
	}{
		{"extend", `{"dropped_headings": ["Abstract", "Zusammenfassung"]}`, []string{"Abstract", "Zusammenfassung"}},
		{"replace", `{"dropped_headings": ["Summary"]}`, []string{"Summary"}},
		{"empty", `{"dropped_headings": []}`, []string{}},
		{"absent", `{"tex": {"ℵ": "\\aleph"}}`, def.DroppedHeadings},
	}
	for _, c := range cases {
		got, err := LoadSymbols(writeOverride(t, c.body))
		if err != nil {
			t.Errorf("%s: error %v", c.name, err)
			continue
		}
		if fmt.Sprint(got.DroppedHeadings) != fmt.Sprint(c.want) {
			t.Errorf("%s: dropped_headings = %v, want %v", c.name, got.DroppedHeadings, c.want)
		}
	}
}

// TestLoadSymbols_MissingPath checks a path that does not exist is an error
// that names the path and satisfies errors.Is(err, fs.ErrNotExist).
func TestLoadSymbols_MissingPath(t *testing.T) {
	p := filepath.Join(t.TempDir(), "no-such-file.json")
	got, err := LoadSymbols(p)
	if err == nil {
		t.Fatal("want an error for a missing override file")
	}
	if got != nil {
		t.Error("a failed load must not return a table")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error %v does not satisfy fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), p) {
		t.Errorf("error %q does not name the file", err)
	}
}

// TestLoadSymbols_BadFiles checks every way an override can be wrong gives an
// error naming the file, and no table.
func TestLoadSymbols_BadFiles(t *testing.T) {
	cases := []struct {
		name, body, mention string
	}{
		{"malformed JSON", `{"tex": {"a": }`, ""},
		{"empty file", ``, ""},
		{"trailing data", `{"tex": {"ℵ": "\\aleph"}} {"x": 1}`, ""},
		{"not an object", `["tex"]`, ""},
		{"unknown section", `{"superscript": {"2": "²"}}`, "superscript"},
		{"empty key", `{"tex": {"": "\\x"}}`, "tex"},
		{"empty tex value", `{"tex": {"ℵ": ""}}`, "tex"},
		{"multi-character superscript key", `{"superscripts": {"ab": "ᵃ"}}`, "superscripts"},
		{"multi-character subscript value", `{"subscripts": {"a": "ₐₐ"}}`, "subscripts"},
		{"empty dropped heading", `{"dropped_headings": ["Abstract", " "]}`, "dropped_headings"},
		{"accent without a command", `{"accents": {"^": "hat"}}`, "accents"},
		{"empty mathvariant key", `{"mathvariants": {"": "\\mathbb"}}`, "mathvariants"},
		{"wrong type", `{"tex": ["a"]}`, ""},
	}
	for _, c := range cases {
		p := writeOverride(t, c.body)
		got, err := LoadSymbols(p)
		if err == nil {
			t.Errorf("%s: want an error", c.name)
			continue
		}
		if got != nil {
			t.Errorf("%s: a failed load must not return a table", c.name)
		}
		if !strings.Contains(err.Error(), p) {
			t.Errorf("%s: error %q does not name the file", c.name, err)
		}
		if c.mention != "" && !strings.Contains(err.Error(), c.mention) {
			t.Errorf("%s: error %q does not mention %q", c.name, err, c.mention)
		}
	}
}

// ExampleLoadSymbols shows loading the embedded table (an empty path means no
// override file) and reading one TeX symbol from it.
func ExampleLoadSymbols() {
	s, err := LoadSymbols("")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s.TeX["τ"])
	// Output: \tau
}
