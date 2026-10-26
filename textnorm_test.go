package crossrefapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

const mathNS = "http://www.w3.org/1998/Math/MathML"

// fixtureTitles returns the "title" array of a captured CrossRef response in
// testdata, named by the DOI with "/" written as "_".
func fixtureTitles(t *testing.T, doiFile string) []string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "crossref-"+doiFile+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Message struct {
			Title []string `json:"title"`
		} `json:"message"`
	}
	if err := json.Unmarshal(src, &r); err != nil {
		t.Fatal(err)
	}
	return r.Message.Title
}

func codes(ws []Warning) []string {
	out := []string{}
	for _, w := range ws {
		out = append(out, w.Code)
	}
	return out
}

// TestPlainText_Fixtures checks real captured titles. The expected strings
// come from the rules the author decided (decision A of crossrefapi/DR-0001),
// not from the code's own output.
func TestPlainText_Fixtures(t *testing.T) {
	cases := []struct {
		file string
		idx  int
		want string
	}{
		// italic with a newline and indent after it: a real space
		{"10.1051_0004-6361_202660273", 0, "Euclid preparation"},
		// layout newlines around <sub>: dropped, so the formula stays whole
		{"10.3847_1538-4357_ae99c4", 0, "Irradiation and Sublimation of Astrophysically Relevant H₂O:NH₃:CO Ices"},
		// one-character italics are glued to their neighbours
		{"10.1021_acs.cgd.6c00364", 0, "Crystal Growth and Characterization of the Ultrahigh-Temperature Substrate Ta1–xHfxC0.5"},
		// a plain space after </sub> is kept
		{"10.1088_1361-648x_ad8697", 0, "Layer-number and strain effects on the structural and electronic properties of PtSe₂ material"},
		// a no-break space is whitespace
		{"10.1007_s00717-026-00622-0", 1, "X-linked juvenile retinoschisis"},
		{"10.1007_s00717-026-00622-0", 0, "X-chromosomale juvenile Retinoschisis"},
		// the title repeats the book title in an italic, between layout newlines
		{"10.1080_00253359.2024.2340221", 0, "Lame Captains and Left-handed Admirals: Amputee officers in Nelson’s navy Lame Captains and Left-handed Admirals: Amputee officers in Nelson’s navy by T. Michals, University of Virginia Press, 2021, $115 (hb), $39.50 (pb) 271 pages, index ISBN 9780813946733"},
		// plain titles are unchanged
		{"10.1007_978-3-319-20325-6", 0, "Data Assimilation"},
		{"10.24272_j.issn.2095-8137.2025.121", 0, "巴马香猪与野猪肾脏发育的动态单核转录图谱"},
	}
	for _, c := range cases {
		titles := fixtureTitles(t, c.file)
		if c.idx >= len(titles) {
			t.Fatalf("%s has %d titles, want index %d", c.file, len(titles), c.idx)
		}
		got, ws := PlainText(titles[c.idx])
		if got != c.want {
			t.Errorf("%s[%d]:\n got  %q\n want %q", c.file, c.idx, got, c.want)
		}
		if len(ws) != 0 {
			t.Errorf("%s[%d]: unexpected warnings %v", c.file, c.idx, ws)
		}
	}
}

// TestPlainText_FixtureMath checks the titles with MathML. Phase 2 keeps a
// formula's text content only (Phase 3 converts it properly), so these assert
// the spacing around the formula and not the formula's own text.
func TestPlainText_FixtureMath(t *testing.T) {
	got, _ := PlainText(fixtureTitles(t, "10.1103_physrevb.85.144303")[0])
	const head = "Line of continuous phase transitions in a three-dimensional U(1) loop model with 1"
	if !strings.HasPrefix(got, head) {
		t.Errorf("a space must be inserted before the formula:\n got  %q\n want prefix %q", got, head)
	}
	if !strings.HasSuffix(got, " current-current interactions") && !strings.Contains(got, " current-current") {
		t.Errorf("a space must be inserted after the formula: %q", got)
	}
	if strings.Contains(got, "with1") || strings.Contains(got, "2current") {
		t.Errorf("formula still abuts the words: %q", got)
	}

	got, _ = PlainText(fixtureTitles(t, "10.1103_physrevd.111.013006")[0])
	if !strings.HasPrefix(got, "τ-lepton pair spin in proton-proton LHC collisions") {
		t.Errorf("a space before a hyphen after a formula must go:\n got %q", got)
	}

	got, _ = PlainText(fixtureTitles(t, "10.1103_physrevb.109.035167")[0])
	if !strings.HasPrefix(got, "Gap collapse and flat band induced by uniaxial strain in ") || strings.Contains(got, "\n") {
		t.Errorf("space and newline before a formula: %q", got)
	}
}

// TestPlainText_AllFixtureTitles checks properties that must hold for every
// captured title: no markup left, tidy whitespace, and idempotence.
func TestPlainText_AllFixtureTitles(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "crossref-*.json"))
	if err != nil || len(files) < 16 {
		t.Fatalf("fixtures: %v, %d files", err, len(files))
	}
	n := 0
	for _, f := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f), "crossref-"), ".json")
		for i, title := range fixtureTitles(t, name) {
			n++
			got, _ := PlainText(title)
			where := name + "[" + string(rune('0'+i)) + "]"
			if strings.ContainsAny(got, "<>") && !strings.Contains(title, "&lt;") && !strings.Contains(title, "&gt;") {
				t.Errorf("%s: markup left: %q", where, got)
			}
			if strings.Contains(got, "mml") {
				t.Errorf("%s: mml left: %q", where, got)
			}
			if got != strings.TrimSpace(got) {
				t.Errorf("%s: untrimmed: %q", where, got)
			}
			for _, r := range got {
				if unicode.IsSpace(r) && r != ' ' {
					t.Errorf("%s: whitespace %U left: %q", where, r, got)
					break
				}
			}
			if strings.Contains(got, "  ") {
				t.Errorf("%s: double space: %q", where, got)
			}
			again, _ := PlainText(got)
			if again != got {
				t.Errorf("%s: not idempotent:\n first  %q\n second %q", where, got, again)
			}
		}
	}
	if n < 16 {
		t.Errorf("only %d titles checked", n)
	}
}

// TestPlainText_Basics covers inputs that are not in a fixture.
func TestPlainText_Basics(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"empty", "", ""},
		{"only whitespace", " \n\t ", ""},
		{"plain unchanged", "A plain title: with punctuation, and (parentheses).", "A plain title: with punctuation, and (parentheses)."},
		{"whitespace collapsed", "  spaced \n out\t title ", "spaced out title"},
		{"predefined entities", "a &lt; b &gt; c &amp; d &quot;e&quot;", `a < b > c & d "e"`},
		{"numeric entity", "caf&#233; &#x41;", "café A"},
		{"entity from the table", "a&nbsp;b&thinsp;c&minus;d", "a b c−d"},
		{"entity from the HTML tables", "caf&eacute;", "café"},
		{"no double decoding", "&amp;lt;", "&lt;"},
		{"raw ampersand", "Tom & Jerry", "Tom & Jerry"},
		{"no-break and thin spaces", "a\u00a0b\u2009c", "a b c"},
		{"strip tag, keep text", "an <b>important</b> title", "an important title"},
		{"attributes ignored", `a <span class="x" id="y">span</span> b`, "a span b"},
		{"nested italics", "<i>a <i>b</i></i> c", "a b c"},
		{"unclosed tag", "<i>abc", "abc"},
		{"prefixed tags", "<jats:italic>Euclid</jats:italic>\n  preparation", "Euclid preparation"},
		{"br is a break", "line one<br/>line two", "line one line two"},
		// scripts
		{"superscript", "x<sup>2</sup>", "x²"},
		{"superscript minus sign", "10<sup>−2</sup>", "10⁻²"},
		{"subscript run", "Ta<sub>1-x</sub>", "Ta₁₋ₓ"},
		{"prefixed subscript", "H<jats:sub>2</jats:sub>O", "H₂O"},
		// whitespace around inline tags (decision A)
		{"layout newline beside sub", "H\n   <sub>2</sub>\n   O", "H₂O"},
		{"layout newline beside sup", "<sup>12</sup>\n   CO(1-0)", "¹²CO(1-0)"},
		{"plain space beside sub is kept", "PtSe<sub>2</sub> material", "PtSe₂ material"},
		{"plain spaces around one-letter italic", "a <i>x</i> b", "a x b"},
		{"newlines around one-letter italic", "a\n  <i>x</i>\n  b", "axb"},
		{"newlines around longer italic", "a\n  <i>xy</i>\n  b", "a xy b"},
		{"one-letter bold is not glued", "a\n  <b>x</b>\n  b", "a x b"},
		{"one-letter prefixed italic is glued", "a\n  <jats:italic>x</jats:italic>\n  b", "axb"},
		// F1 (decision A amended): a deliberate typographic space beside the layout newline keeps a space
		{"thin spaces around a one-letter italic keep their spaces", "4\u2009&lt;\u2009\n  <i>M</i>\n  \u2009&lt;\u20098", "4 < M < 8"},
		{"en spaces around one-letter italics", "<i>k</i>\n  \u2004&lt;\u20040.5\u2005\n  <i>h</i>\n  /Mpc", "k < 0.5 h/Mpc"},
		{"a typographic space beside a subscript keeps one space", "H\u2009\n  <sub>2</sub>", "H ₂"},
		{"a no-break space beside a layout newline keeps one space", "a\u00a0\n  <i>x</i>", "a x"},
		{"plain space, tab and newline are still layout", "a \n \t <i>x</i>", "ax"},
		// F2 (decision A amended): a layout newline between a formula and closing punctuation goes
		{"formula then comma", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  , and y", "x, and y"},
		{"formula then period", `y <mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  . Next", "y x. Next"},
		{"formula then semicolon", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  ; z", "x; z"},
		{"formula then closing bracket", "(" + `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  )", "(x)"},
		{"formula then closing square bracket", "[" + `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  ]", "[x]"},
		{"opening bracket then formula", "(\n  " + `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  )", "(x)"},
		{"opening square bracket then formula", "see [\n  " + `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "]", "see [x]"},
		{"a plain space before a comma after a formula stays", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math> , y`, "x , y"},
		{"a layout newline before a comma with no formula stays a space", "word\n  , y", "word , y"},
		{"a colon is not closing punctuation here", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  : y", "x : y"},
		{"a bracket after a formula but before text is not an opener", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + "\n  (y)", "x (y)"},
		{"a plain space after an opening bracket before a formula stays", "( " + `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>` + ")", "( x)"},
		{"opening bracket and a word", "(\n  word", "( word"},
		{"glued around sub and sup", "10<sup>6</sup>\n M\n <sub>2</sub>\n pc", "10⁶M₂pc"},
		// MathML
		{"math abutting words, both sides", `with<mml:math xmlns:mml="` + mathNS + `"><mml:mn>1</mml:mn></mml:math>current`, "with 1 current"},
		{"math recognized by namespace, not prefix", `with<a:math xmlns:a="` + mathNS + `"><a:mn>1</a:mn></a:math>current`, "with 1 current"},
		{"math layout whitespace dropped", `near <mml:math xmlns:mml="` + mathNS + `">` + "\n <mml:mi>J</mml:mi>\n <mml:mo>/</mml:mo>\n <mml:mi>ψ</mml:mi>\n </mml:math> photoproduction", "near J/ψ photoproduction"},
		{"space before hyphen after a formula goes", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>τ</mml:mi></mml:math> -lepton`, "τ-lepton"},
		{"space before hyphen without a formula stays", "word -lepton", "word -lepton"},
		{"space before a U+2010 hyphen after a formula goes", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>τ</mml:mi></mml:math> ‐lepton`, "τ‐lepton"},
		{"hyphen rule applies only right after the formula", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math> y -z`, "x y -z"},
		{"no space inserted beside a bracket", `U(1)<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>(y)`, "U(1)x(y)"},
		{"one-letter em is glued", "a\n  <em>x</em>\n  b", "axb"},
		{"carriage return entity is a newline", "a&#13;<sub>2</sub>", "a₂"},
		{"unclosed br", "line one<br>line two", "line one line two"},
		{"empty script has no effect", "a\n<sub></sub>\nb", "a b"},
		{"space before an en dash after a formula stays", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math> – y`, "x – y"},
	}
	for _, c := range cases {
		got, ws := PlainText(c.in)
		if got != c.want {
			t.Errorf("%s:\n in   %q\n got  %q\n want %q", c.name, c.in, got, c.want)
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.name, codes(ws))
		}
	}
}

// TestPlainText_Warnings covers the cases that warn.
func TestPlainText_Warnings(t *testing.T) {
	// a subscript the table cannot map keeps its text and warns
	got, ws := PlainText("Hf<sub>y</sub>C")
	if got != "HfyC" {
		t.Errorf("unmapped subscript: got %q, want %q", got, "HfyC")
	}
	if len(ws) != 1 || ws[0].Code != WarnScriptUnmapped || ws[0].Fragment != "<sub>y</sub>" {
		t.Errorf("unmapped subscript: warnings %+v", ws)
	}
	// partly mappable still warns, and keeps the whole text
	got, ws = PlainText("x<sup>2q</sup>")
	if got != "x2q" || len(ws) != 1 || ws[0].Code != WarnScriptUnmapped {
		t.Errorf("partly mappable superscript: got %q, warnings %+v", got, ws)
	}
	// the glued case from decision A: an unmappable subscript is kept as text and glued
	got, ws = PlainText("10<sup>6</sup>\n M\n <sub>⊙</sub>\n pc")
	if got != "10⁶M⊙pc" || len(ws) != 1 || ws[0].Code != WarnScriptUnmapped || ws[0].Fragment != "<sub>⊙</sub>" {
		t.Errorf("glued unmappable subscript: got %q, warnings %+v", got, ws)
	}
	// a raw "<" is not markup: text is kept, with one parse-fallback warning
	got, ws = PlainText("when a < b and c")
	if got != "when a < b and c" {
		t.Errorf("raw <: got %q", got)
	}
	if len(ws) != 1 || ws[0].Code != WarnParseFallback {
		t.Errorf("raw <: warnings %+v", ws)
	}
	// the fallback still strips real tags and decodes entities
	got, ws = PlainText("x < <i>y</i> &amp; z")
	if got != "x < y & z" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
		t.Errorf("fallback with tags: got %q, warnings %+v", got, ws)
	}
	// clean input has no warnings (not even an empty non-nil slice is required)
	if _, ws = PlainText("clean <i>title</i>"); len(ws) != 0 {
		t.Errorf("clean input warned: %+v", ws)
	}
}

// TestPlainText_NestedWarnings checks a warning from inside a script reaches
// the caller once, and that an italic around a script does not count it twice.
func TestPlainText_NestedWarnings(t *testing.T) {
	_, ws := PlainText("x<sub>2<sup>q</sup></sub>")
	if len(ws) != 2 {
		t.Errorf("nested scripts: want 2 warnings (inner and outer), got %+v", ws)
	}
	got, ws := PlainText("a <i><sub>q</sub></i> b")
	if got != "a q b" || len(ws) != 1 {
		t.Errorf("italic around an unmapped script: got %q, warnings %+v", got, ws)
	}
}

// TestPlainText_FallbackDetails checks what the parse fallback reports and
// strips.
func TestPlainText_FallbackDetails(t *testing.T) {
	in := "a < b </i> c"
	got, ws := PlainText(in)
	if got != "a < b c" {
		t.Errorf("fallback strips a closing tag: got %q", got)
	}
	if len(ws) != 1 || ws[0].Fragment != in {
		t.Errorf("fallback warning must carry the fragment: %+v", ws)
	}
	if w := ws[0].String(); w != WarnParseFallback+": "+in {
		t.Errorf("Warning.String() = %q", w)
	}
	// an end tag that would close the parser's own wrapper is not well-formed either
	got, ws = PlainText("abc</root>def")
	if got != "abcdef" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
		t.Errorf("wrapper end tag: got %q, warnings %+v", got, ws)
	}
	// an end tag with no start tag is not well-formed: text kept, with a warning
	got, ws = PlainText("abc</i>def")
	if got != "abcdef" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
		t.Errorf("stray end tag: got %q, warnings %+v", got, ws)
	}
}

// TestPlainText_TablePrecedence checks an entity in the symbol table wins over
// the HTML entity tables, so an override file can change what an entity means.
func TestPlainText_TablePrecedence(t *testing.T) {
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.Entities["eacute"] = "E"
	got, _ := plainText("caf&eacute; &nbsp;x", sym)
	if got != "cafE x" {
		t.Errorf("got %q, want the table's value for &eacute;", got)
	}
}
