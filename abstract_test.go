package crossrefapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// sanitizerTags is RDM's allowed HTML tag list, copied from marshmallow_utils
// 0.15.2 (ALLOWED_HTML_TAGS), the same in the CaltechAUTHORS and CaltechTHESIS
// environments. HTMLAbstract output must stay inside it.
var sanitizerTags = map[string]bool{
	"a": true, "abbr": true, "acronym": true, "b": true, "blockquote": true, "br": true,
	"code": true, "col": true, "colgroup": true, "div": true, "table": true, "tbody": true,
	"tfoot": true, "thead": true, "td": true, "th": true, "tr": true, "em": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "i": true, "li": true,
	"ol": true, "p": true, "pre": true, "s": true, "span": true, "strike": true,
	"strong": true, "sub": true, "sup": true, "u": true, "ul": true,
}

var htmlTagRE = regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9]*)`)

// fixtureAbstracts returns every captured abstract, keyed by file name.
func fixtureAbstracts(t *testing.T) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "crossref-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var r struct {
			Message struct {
				Abstract string `json:"abstract"`
			} `json:"message"`
		}
		if err := json.Unmarshal(src, &r); err != nil {
			t.Fatal(err)
		}
		if r.Message.Abstract != "" {
			out[strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f), "crossref-"), ".json")] = r.Message.Abstract
		}
	}
	return out
}

// balanced reports whether every non-void tag in html is closed in order.
func balanced(html string) bool {
	var stack []string
	for _, m := range regexp.MustCompile(`<(/?)([A-Za-z][A-Za-z0-9]*)[^>]*>`).FindAllStringSubmatch(html, -1) {
		name := m[2]
		if name == "br" {
			continue
		}
		if m[1] == "" {
			stack = append(stack, name)
			continue
		}
		if len(stack) == 0 || stack[len(stack)-1] != name {
			return false
		}
		stack = stack[:len(stack)-1]
	}
	return len(stack) == 0
}

// TestHTMLAbstract_Golden checks one real abstract in full. The expected text
// is the source text with the rules applied by hand: the leading "Abstract"
// heading dropped, one paragraph, and the real "(4 < M < 8)" keeping its thin
// spaces beside the layout newlines around the one-letter italic M (F1).
func TestHTMLAbstract_Golden(t *testing.T) {
	want := "<p>Earthquake onsets provide a unique opportunity to study physical rupture processes " +
		"because they are more easily observable than later rupture stages. Despite this relative " +
		"simplicity, the observational basis for rupture onsets is unclear. Numerous reports of " +
		"evidence for magnitude{h}dependent rupture onsets (which imply deterministic rupture " +
		"behavior) stand in contradiction to a large body of physics{h}based rupture modeling " +
		"efforts, which are mostly based on inherently nondeterministic principles. Here we make " +
		"use of the abundance of short{h}distance recordings available today; a magnitude " +
		"dependency of onsets should appear most prominently in such recordings. We use a simple " +
		"method to demonstrate that all ruptures in the studied magnitude range (4 &lt; <i>M</i> &lt; 8) " +
		"share a universal initial rupture behavior and discuss ensuing implications for physical " +
		"rupture processes and earthquake early warning.</p>"
	want = strings.ReplaceAll(want, "{h}", "\u2010")
	got, ws := HTMLAbstract(fixtureAbstracts(t)["10.1002_2016GL070081"])
	if got != want {
		t.Errorf("golden abstract differs:\n got  %q\n want %q", got, want)
	}
	if len(ws) != 0 {
		t.Errorf("unexpected warnings %v", ws)
	}
}

// TestHTMLAbstract_Fixtures checks the headings, formulas and dropped text of
// real abstracts.
func TestHTMLAbstract_Fixtures(t *testing.T) {
	cases := []struct {
		file     string
		prefix   string
		contains []string
		absent   []string
	}{
		{"10.1103_physrevb.109.035167", "", []string{`\(1T-\mathrm{TaS}_{2}\)`, `\(d_{z^{2}}\)`},
			[]string{"Published by the American Physical Society", "2024", "<b></b>", "jats", "xmlns"}},
		{"10.1103_physrevlett.134.141902", "", []string{`\(\mathbb{Z}_{N}\)`, `\(\mathrm{SU}(N)\)`, `\(d=2\), only the confined`, `up to (\(N-1\)) photons`}, []string{`\) ,`, `( \(`, `\) )`}},
		{"10.1103_sh13-grfw", "", []string{`\(J/\psi\)`, `\(\overline{D}^{0}\mathrm{\Lambda}_{c}^{+}\)`, `\(\gamma p\to\overline{D}^{(*)0}\mathrm{\Lambda}_{c}^{+}\)`, `GlueX, \(J/\psi-007\), and CLAS`}, []string{`\) ,`}},
		{"10.1103_physrevd.111.013006", "", []string{`\(\tau\)`}, nil},
		{"10.1007_s41114-018-0012-9", "", []string{`\(90\%\)`, `\(20~\mathrm{deg}^{2}\)`, `\(\sim 2\)`}, []string{"$$", "tex-math", "alternatives"}},
		{"10.1007_s00332-025-10238-4", "", []string{`\(n-1\)`}, []string{"$$"}},
		// the AAS formula whose TeX is blank still converts, and is a prescript
		{"10.3847_1538-4357_ae99c4", "", []string{`\({}_{4}^{+}\)`, `\({}_{4}^{+}\)), formaldehyde`}, []string{"tex-math", `\) )`}},
		// real markup: italics around single letters, thin and em spaces, a numeric space
		{"10.1051_0004-6361_202660273", "<p>", []string{"(<i>k</i> &lt; 0.5 <i>h</i>/Mpc)", "cloned <i>Euclid</i> large", "Using 10 000 cloned"}, []string{"jats", "\u2004", "\u2005"}},
		// the first "Abstract" heading, alone in a section, is dropped; the structured headings stay
		{"10.1007_s11695-024-07110-x", "<p><b>Background</b></p><p>Bariatric surgery", []string{"<p><b>Aim</b></p>", "<p><b>Graphical Abstract</b></p>"}, []string{"<b>Abstract</b></p><p><b>Background"}},
		{"10.2174_0118761429257350231212093136", "<p><b>Background::</b></p><p>Amplification", []string{"<p><b>Methods::</b></p>"}, nil},
		// a German heading is not in the list, so it stays
		{"10.1007_s00717-026-00622-0", "<p><b>Zusammenfassung</b></p><p>Die X-chromosomale", nil, nil},
	}
	abs := fixtureAbstracts(t)
	for _, c := range cases {
		src, ok := abs[c.file]
		if !ok {
			t.Errorf("%s: no abstract fixture", c.file)
			continue
		}
		got, _ := HTMLAbstract(src)
		if c.prefix != "" && !strings.HasPrefix(got, c.prefix) {
			t.Errorf("%s: starts %q, want prefix %q", c.file, got[:min(len(got), len(c.prefix)+20)], c.prefix)
		}
		for _, w := range c.contains {
			if !strings.Contains(got, w) {
				t.Errorf("%s: output lacks %q", c.file, w)
			}
		}
		for _, w := range c.absent {
			if strings.Contains(got, w) {
				t.Errorf("%s: output has %q", c.file, w)
			}
		}
	}
}

// TestHTMLAbstract_AllFixtures checks properties that must hold for every real
// abstract: only sanitizer-allowed tags, no MathML or JATS left, balanced and
// tidy, no leading "Abstract" heading, formulas delimited, and idempotent.
func TestHTMLAbstract_AllFixtures(t *testing.T) {
	abs := fixtureAbstracts(t)
	if len(abs) < 12 {
		t.Fatalf("only %d abstract fixtures", len(abs))
	}
	for name, src := range abs {
		got, _ := HTMLAbstract(src)
		if got == "" {
			t.Errorf("%s: empty output", name)
			continue
		}
		for _, m := range htmlTagRE.FindAllStringSubmatch(got, -1) {
			if !sanitizerTags[m[1]] {
				t.Errorf("%s: tag <%s> is not in the sanitizer's list", name, m[1])
			}
		}
		for _, bad := range []string{"mml", "jats", "xmlns", "tex-math", "<math", "$$"} {
			if strings.Contains(got, bad) {
				t.Errorf("%s: output contains %q", name, bad)
			}
		}
		if !balanced(got) {
			t.Errorf("%s: tags are not balanced", name)
		}
		if strings.Count(got, `\(`) != strings.Count(got, `\)`) {
			t.Errorf("%s: unbalanced formula delimiters", name)
		}
		if strings.ContainsAny(got, "\n\r\t") || strings.Contains(got, "  ") || got != strings.TrimSpace(got) {
			t.Errorf("%s: untidy whitespace: %q", name, got[:min(len(got), 120)])
		}
		if regexp.MustCompile(`\\\) [,.;)\]]|[(\[] \\\(`).MatchString(got) {
			t.Errorf("%s: a space is left between a formula and punctuation: %q", name, regexp.MustCompile(`.{0,20}(\\\) [,.;)\]]|[(\[] \\\().{0,20}`).FindString(got))
		}
		for _, empty := range []string{"<p></p>", "<p><b></b></p>"} {
			if strings.Contains(got, empty) {
				t.Errorf("%s: empty block %s", name, empty)
			}
		}
		if strings.HasPrefix(got, "<p><b>Abstract</b></p>") {
			t.Errorf("%s: starts with the Abstract heading", name)
		}
		again, _ := HTMLAbstract(got)
		if again != got {
			t.Errorf("%s: not idempotent:\n first  %q\n second %q", name, got[:min(len(got), 160)], again[:min(len(again), 160)])
		}
	}
}

// TestHTMLAbstract_Basics covers inputs that are not in a fixture.
func TestHTMLAbstract_Basics(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"empty", "", ""},
		{"only whitespace", " \n ", ""},
		{"plain text is escaped", "Tom &amp; Jerry &lt;3", "Tom &amp; Jerry &lt;3"},
		{"two paragraphs", "<jats:p>One</jats:p><jats:p>Two</jats:p>", "<p>One</p><p>Two</p>"},
		{"layout between blocks", "\n  <jats:p>\n  One\n </jats:p>\n  <jats:p>Two</jats:p>\n", "<p>One</p><p>Two</p>"},
		{"empty paragraph is skipped", "<jats:p>One</jats:p><jats:p>  </jats:p>", "<p>One</p>"},
		{"section is transparent", "<jats:sec><jats:p>One</jats:p></jats:sec>", "<p>One</p>"},
		// headings (decision B)
		{"heading becomes a bold paragraph", "<jats:p>x</jats:p><jats:title>Methods</jats:title><jats:p>y</jats:p>", "<p>x</p><p><b>Methods</b></p><p>y</p>"},
		{"leading Abstract dropped", "<jats:title>Abstract</jats:title><jats:p>y</jats:p>", "<p>y</p>"},
		{"leading ABSTRACT dropped", "<jats:title>ABSTRACT</jats:title><jats:p>y</jats:p>", "<p>y</p>"},
		{"leading Abstract: dropped", "<jats:title> Abstract: </jats:title><jats:p>y</jats:p>", "<p>y</p>"},
		{"leading heading not in the list stays", "<jats:title>Summary</jats:title><jats:p>y</jats:p>", "<p><b>Summary</b></p><p>y</p>"},
		{"only the first heading is dropped", "<jats:title>Abstract</jats:title><jats:title>Abstract</jats:title><jats:p>y</jats:p>", "<p><b>Abstract</b></p><p>y</p>"},
		{"Abstract after content stays", "<jats:p>x</jats:p><jats:title>Abstract</jats:title>", "<p>x</p><p><b>Abstract</b></p>"},
		{"Abstract inside a section after the first title stays", "<jats:title>Background</jats:title><jats:sec><jats:title>Abstract</jats:title></jats:sec><jats:p>y</jats:p>", "<p><b>Background</b></p><p><b>Abstract</b></p><p>y</p>"},
		{"empty heading is skipped", "<jats:title/><jats:p>y</jats:p>", "<p>y</p>"},
		{"an empty first heading does not use up the drop", "<jats:title/><jats:title>Abstract</jats:title><jats:p>y</jats:p>", "<p>y</p>"},
		{"heading with italic", "<jats:title>The <jats:italic>E. coli</jats:italic> case</jats:title>", "<p><b>The <i>E. coli</i> case</b></p>"},
		// text and elements
		{"italic and bold", "a <jats:italic>b</jats:italic> and <jats:bold>c</jats:bold> d", "a <i>b</i> and <b>c</b> d"},
		{"sub and sup stay tags", "H<jats:sub>2</jats:sub>O and 10<jats:sup>6</jats:sup>", "H<sub>2</sub>O and 10<sup>6</sup>"},
		{"layout beside sup is dropped", "<jats:sup>12</jats:sup>\n   CO(1-0)", "<sup>12</sup>CO(1-0)"},
		{"plain space beside sub is kept", "PtSe<jats:sub>2</jats:sub> material", "PtSe<sub>2</sub> material"},
		{"one-letter italic is glued", "a\n  <jats:italic>x</jats:italic>\n  b", "a<i>x</i>b"},
		{"longer italic keeps its spaces", "a\n  <jats:italic>xy</jats:italic>\n  b", "a <i>xy</i> b"},
		{"one-letter em is glued", "a\n  <em>x</em>\n  b", "a<em>x</em>b"},
		{"empty inline element vanishes", "a<i></i>b <jats:sub> </jats:sub>c", "ab c"},
		{"other inline tags pass through", "<u>a</u> <s>b</s> <strike>c</strike> <code>d</code> <span>e</span>", "<u>a</u> <s>b</s> <strike>c</strike> <code>d</code> <span>e</span>"},
		{"block tags pass through", "<h1>a</h1><h2>b</h2><h3>c</h3><h4>d</h4><h5>e</h5><blockquote>q</blockquote><pre>r</pre><div>d</div>", "<h1>a</h1><h2>b</h2><h3>c</h3><h4>d</h4><h5>e</h5><blockquote>q</blockquote><pre>r</pre><div>d</div>"},
		{"ordered list", "<ol><li>x</li></ol>", "<ol><li>x</li></ol>"},
		{"table sections", "<table><thead><tr><th>h</th></tr></thead><tbody><tr><td>b</td></tr></tbody><tfoot><tr><td>f</td></tr></tfoot></table>", "<table><thead><tr><th>h</th></tr></thead><tbody><tr><td>b</td></tr></tbody><tfoot><tr><td>f</td></tr></tfoot></table>"},
		{"formula before a heading counts as content", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math><jats:title>Abstract</jats:title>`, `\(x\)<p><b>Abstract</b></p>`},
		{"display formula with tex-math alone", `<jats:disp-formula><jats:tex-math>$$x$$</jats:tex-math></jats:disp-formula>`, `\(x\)`},
		{"display formula with MathML", `<jats:disp-formula><mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math></jats:disp-formula>`, `\(x\)`},
		{"supplementary material with its own text is dropped", "<jats:p>a</jats:p><jats:supplementary-material>Dataset at X</jats:supplementary-material>", "<p>a</p>"},
		{"layout beside sub is dropped", "H\n  <jats:sub>2</jats:sub>\n  O", "H<sub>2</sub>O"},
		{"empty list item is kept", "<ul><li></li></ul>", "<ul><li></li></ul>"},
		{"a paragraph holding only a formula is kept", `<jats:p><mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math></jats:p>`, `<p>\(x\)</p>`},
		{"span attributes on a paragraph are dropped", `<p colspan="2" rowspan="3">a</p>`, "<p>a</p>"},
		{"alternatives with tex-math alone", `<jats:alternatives><jats:tex-math>$$x$$</jats:tex-math></jats:alternatives>`, `\(x\)`},
		{"tex-math two levels down", `<jats:inline-formula><jats:alternatives><jats:tex-math>$$x$$</jats:tex-math></jats:alternatives></jats:inline-formula>`, `\(x\)`},
		{"small caps unwrapped", "H <jats:sc>I</jats:sc> gas", "H I gas"},
		{"unknown element unwrapped", "a <jats:foo>b</jats:foo> c", "a b c"},
		{"supplementary material dropped with its content", "<jats:p>a</jats:p><jats:supplementary-material><jats:permissions><jats:copyright-statement>Published by X</jats:copyright-statement><jats:copyright-year>2024</jats:copyright-year></jats:permissions></jats:supplementary-material>", "<p>a</p>"},
		{"permissions dropped with their content", "a <jats:permissions>(c) Y</jats:permissions>b", "a b"},
		{"script and style dropped with their content", "a<script>alert(1)</script>b<style>p{}</style>c", "abc"},
		{"bare HTML tags are kept", "<p>a <em>b</em> <strong>c</strong></p>", "<p>a <em>b</em> <strong>c</strong></p>"},
		{"br is a break", "a<br>b", "a<br>b"},
		{"br with layout", "a\n  <br/>\n  b", "a<br>b"},
		{"lists", "<ul>\n <li>a</li>\n <li>b</li>\n</ul>", "<ul><li>a</li><li>b</li></ul>"},
		{"tables", "<table>\n <tr>\n  <th colspan=\"2\">h</th>\n </tr>\n <tr><td>a</td><td></td></tr>\n</table>", `<table><tr><th colspan="2">h</th></tr><tr><td>a</td><td></td></tr></table>`},
		{"nbsp and thin space are spaces", "a&nbsp;b\u2009c", "a b c"},
		{"entities", "a &lt; b &amp; c &gt; d", "a &lt; b &amp; c &gt; d"},
		// attributes
		{"attributes dropped", `<p class="x" style="color:red" onclick="e()" lang="en">a</p>`, "<p>a</p>"},
		{"link keeps href and title", `<a href="https://x.org/a?b=1&amp;c=2" title="T" onclick="e()" target="_blank">l</a>`, `<a href="https://x.org/a?b=1&amp;c=2" title="T">l</a>`},
		{"http link", `<a href="http://x.org">h</a>`, `<a href="http://x.org">h</a>`},
		{"upper-case scheme", `<a href="HTTP://X.ORG">h</a>`, `<a href="HTTP://X.ORG">h</a>`},
		{"link title is escaped", `<a href="/x" title="say &quot;hi&quot; &lt;b&gt; &amp;">l</a>`, `<a href="/x" title="say &quot;hi&quot; &lt;b&gt; &amp;">l</a>`},
		{"link with tab in the address is unwrapped", "<a href=\"ja&#9;vascript:alert(1)\">l</a>", "l"},
		{"acronym keeps title", `<acronym title="X">x</acronym>`, `<acronym title="X">x</acronym>`},
		{"long colspan dropped", `<table><tr><td colspan="12345">a</td></tr></table>`, `<table><tr><td>a</td></tr></table>`},
		{"mailto link", `<a href="mailto:a@b.org">m</a>`, `<a href="mailto:a@b.org">m</a>`},
		{"relative link", `<a href="/x/y">r</a>`, `<a href="/x/y">r</a>`},
		{"javascript link is unwrapped", `<a href="javascript:alert(1)">l</a>`, "l"},
		{"mixed-case javascript link is unwrapped", `<a href="JaVaScRiPt:alert(1)">l</a>`, "l"},
		{"data link is unwrapped", `<a href="data:text/html,x">l</a>`, "l"},
		{"link with no href is unwrapped", `<a name="top">l</a>`, "l"},
		{"abbr keeps title", `<abbr title="Hubble Space Telescope" id="z">HST</abbr>`, `<abbr title="Hubble Space Telescope">HST</abbr>`},
		{"bad colspan dropped", `<table><tr><td colspan="x" rowspan="3">a</td></tr></table>`, `<table><tr><td rowspan="3">a</td></tr></table>`},
		// formulas (decision E: MathML always converted, tex-math ignored)
		{"formula between words", `with<mml:math xmlns:mml="` + mathNS + `"><mml:mn>1</mml:mn></mml:math>current`, `with \(1\) current`},
		{"formula by namespace, not prefix", `with<a:math xmlns:a="` + mathNS + `"><a:mi>x</a:mi></a:math>current`, `with \(x\) current`},
		{"formula then hyphen", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>τ</mml:mi></mml:math> -lepton`, `\(\tau\)-lepton`},
		{"formula with a less-than sign", `<mml:math xmlns:mml="` + mathNS + `"><mml:mi>a</mml:mi><mml:mo>&lt;</mml:mo><mml:mi>b</mml:mi></mml:math>`, `\(a&lt;b\)`},
		{"formula with an ampersand", `<mml:math xmlns:mml="` + mathNS + `"><mml:mtext>a &amp; b</mml:mtext></mml:math>`, `\(\text{a \&amp; b}\)`},
		{"alternatives keeps the MathML only", `<jats:inline-formula><jats:alternatives><jats:tex-math>$$90\%$$</jats:tex-math><mml:math xmlns:mml="` + mathNS + `"><mml:mn>90</mml:mn><mml:mo>%</mml:mo></mml:math></jats:alternatives></jats:inline-formula>`, `\(90\%\)`},
		{"blank tex-math beside MathML", `<jats:inline-formula><jats:tex-math> </jats:tex-math><mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math></jats:inline-formula>`, `\(x\)`},
		{"empty formula vanishes", `a <mml:math xmlns:mml="` + mathNS + `"></mml:math> b`, `a b`},
		// tex-math used only when there is no MathML at all (an addition to decision E, synthetic)
		{"tex-math alone: $$ delimiters", `<jats:inline-formula><jats:tex-math>$$x^2$$</jats:tex-math></jats:inline-formula>`, `\(x^2\)`},
		{"tex-math alone: $ delimiters", `<jats:inline-formula><jats:tex-math>$x$</jats:tex-math></jats:inline-formula>`, `\(x\)`},
		{"tex-math alone: \\( delimiters", `<jats:inline-formula><jats:tex-math>\(x\)</jats:tex-math></jats:inline-formula>`, `\(x\)`},
		{"tex-math alone: \\[ delimiters", `<jats:inline-formula><jats:tex-math>\[x\]</jats:tex-math></jats:inline-formula>`, `\(x\)`},
		{"tex-math alone: no delimiters", `<jats:inline-formula><jats:tex-math>x</jats:tex-math></jats:inline-formula>`, `\(x\)`},
		{"tex-math alone: escaped", `<jats:inline-formula><jats:tex-math>$a&lt;b$</jats:tex-math></jats:inline-formula>`, `\(a&lt;b\)`},
		{"tex-math alone: blank", `a <jats:inline-formula><jats:tex-math>  </jats:tex-math></jats:inline-formula> b`, `a b`},
	}
	for _, c := range cases {
		got, ws := HTMLAbstract(c.in)
		if got != c.want {
			t.Errorf("%s:\n in   %q\n got  %q\n want %q", c.name, c.in, got, c.want)
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.name, ws)
		}
	}
}

// TestHTMLAbstract_Warnings covers what warns.
func TestHTMLAbstract_Warnings(t *testing.T) {
	// a raw "<" is not markup: text kept, escaped, with one parse-fallback warning
	got, ws := HTMLAbstract("when a < b <jats:p>and</jats:p> c")
	if got != "when a &lt; b and c" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
		t.Errorf("raw <: got %q, warnings %+v", got, ws)
	}
	// a formula that cannot be converted warns, and is kept as text
	got, ws = HTMLAbstract(`<jats:p>x <mml:math xmlns:mml="` + mathNS + `"><mml:mtable><mml:mtr><mml:mtd><mml:mi>a</mml:mi></mml:mtd></mml:mtr></mml:mtable></mml:math> y</jats:p>`)
	if got != `<p>x \(\text{a}\) y</p>` || len(ws) != 1 || ws[0].Code != WarnMathUnconverted || ws[0].Fragment != "mtable" {
		t.Errorf("unconverted formula: got %q, warnings %+v", got, ws)
	}
}

// TestHTMLAbstract_DroppedHeadingsFromTable checks the heading list comes from
// the symbol table, so the override file can extend or empty it.
func TestHTMLAbstract_DroppedHeadingsFromTable(t *testing.T) {
	src := "<jats:title>Abstract</jats:title><jats:p>y</jats:p>"
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.DroppedHeadings = nil
	if got, _ := htmlAbstract(src, sym); got != "<p><b>Abstract</b></p><p>y</p>" {
		t.Errorf("empty list drops nothing: got %q", got)
	}
	sym.DroppedHeadings = []string{"Summary", "  Zusammenfassung:  ", "Plain   language summary"}
	for in, want := range map[string]string{
		"<jats:title>summary</jats:title><jats:p>y</jats:p>":                "<p>y</p>",
		"<jats:title>Zusammenfassung</jats:title><jats:p>y</jats:p>":        "<p>y</p>",
		"<jats:title>Abstract</jats:title><jats:p>y</jats:p>":               "<p><b>Abstract</b></p><p>y</p>",
		"<jats:title>Plain language summary</jats:title><jats:p>y</jats:p>": "<p>y</p>",
	} {
		if got, _ := htmlAbstract(in, sym); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

// TestParse_Attributes checks the parser keeps attributes by local name and
// leaves out namespace declarations.
func TestParse_Attributes(t *testing.T) {
	sym, _ := DefaultSymbols()
	root, err := parseFragment(`<a href="/x" xmlns:q="urn:q" q:Title="T" xmlns="urn:d">l</a>`, sym)
	if err != nil {
		t.Fatal(err)
	}
	a := root.kids[0].kids[0] // the synthetic <root> wrapper comes first
	if len(a.attrs) != 2 || a.attrs["href"] != "/x" || a.attrs["title"] != "T" {
		t.Errorf("attrs = %v, want href and title only", a.attrs)
	}
}
