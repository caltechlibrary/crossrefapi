package crossrefapi

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

// robustLimit is far above the linear time (well under a second, a few seconds under
// the race detector on a busy machine) and far below the superlinear time these
// tests guard against (a minute or more).
const robustLimit = 20 * time.Second

// within runs f and fails the test if it takes longer than limit. A superlinear
// algorithm shows up here as a timeout instead of a hang.
func within(t *testing.T, limit time.Duration, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(limit):
		t.Fatalf("%s: still running after %v", what, limit)
	}
}

// TestRobust_NestedItalics checks nesting does not make the work grow
// exponentially (the one-letter-italic test once converted the children twice
// at every level).
func TestRobust_NestedItalics(t *testing.T) {
	for _, depth := range []int{30, 60, 120} {
		in := strings.Repeat("<i>", depth) + "x" + strings.Repeat("</i>", depth)
		within(t, robustLimit, fmt.Sprintf("PlainText, %d nested italics", depth), func() {
			if got, _ := PlainText(in); got != "x" {
				t.Errorf("depth %d: PlainText = %q", depth, got)
			}
		})
		within(t, robustLimit, fmt.Sprintf("HTMLAbstract, %d nested italics", depth), func() {
			got, _ := HTMLAbstract(in)
			if strings.Count(got, "<i>") != depth || !strings.Contains(got, "x") {
				t.Errorf("depth %d: HTMLAbstract lost an italic: %.60q", depth, got)
			}
		})
	}
}

// TestRobust_LargeFormula checks a formula of many tokens converts in time
// proportional to its size.
func TestRobust_LargeFormula(t *testing.T) {
	in := `<mml:math xmlns:mml="` + mathNS + `">` + strings.Repeat("<mml:mi>α</mml:mi>", 30000) + `</mml:math>`
	within(t, robustLimit, "HTMLAbstract, 30000 formula tokens", func() {
		got, _ := HTMLAbstract(in)
		if strings.Count(got, `\alpha`) != 30000 {
			t.Errorf("lost tokens: %d", strings.Count(got, `\alpha`))
		}
	})
	within(t, robustLimit, "PlainText, 30000 formula tokens", func() {
		if got, _ := PlainText(in); len([]rune(got)) != 30000 {
			t.Errorf("PlainText length %d", len([]rune(got)))
		}
	})
}

// TestRobust_FallbackLinear checks the fallback path (reached by a raw "<") is
// linear: a long run of "<a" with no closing ">" once rescanned the rest of the
// text for every one.
func TestRobust_FallbackLinear(t *testing.T) {
	in := "> < " + strings.Repeat("<a", 2000000)
	within(t, robustLimit, "fallback over 2000000 '<a'", func() {
		got, ws := PlainText(in)
		if len(ws) != 1 || ws[0].Code != WarnParseFallback || !strings.HasPrefix(got, "> <") {
			t.Errorf("got %.40q, warnings %v", got, ws)
		}
	})
}

// TestRobust_DepthLimit checks very deep nesting is refused cheaply, as a
// parse-fallback with the text kept, while ordinary nesting is converted.
func TestRobust_DepthLimit(t *testing.T) {
	deep := strings.Repeat("<span>", 100000) + "x" + strings.Repeat("</span>", 100000)
	within(t, robustLimit, "HTMLAbstract, 100000 nested spans", func() {
		got, ws := HTMLAbstract(deep)
		if got != "x" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
			t.Errorf("got %.40q, warnings %d", got, len(ws))
		}
	})
	within(t, robustLimit, "PlainText, 100000 nested bold", func() {
		got, ws := PlainText(strings.NewReplacer("span", "b").Replace(deep))
		if got != "x" || len(ws) != 1 || ws[0].Code != WarnParseFallback {
			t.Errorf("got %.40q, warnings %d", got, len(ws))
		}
	})
	ok := strings.Repeat("<span>", 200) + "x" + strings.Repeat("</span>", 200)
	if got, ws := HTMLAbstract(ok); strings.Count(got, "<span>") != 200 || len(ws) != 0 {
		t.Errorf("200 nested spans should convert: %d spans, warnings %v", strings.Count(got, "<span>"), ws)
	}
	over := strings.Repeat("<span>", 300) + "x" + strings.Repeat("</span>", 300)
	if got, ws := HTMLAbstract(over); got != "x" || len(ws) != 1 {
		t.Errorf("300 nested spans should fall back: %.40q, %v", got, ws)
	}
}

// TestRobust_FallbackDropsScripts checks the fallback leaves out script and
// style content the way the normal path does.
func TestRobust_FallbackDropsScripts(t *testing.T) {
	got, _ := PlainText("x < y <script>alert(1)</script> z <STYLE type=\"a\">p{}</STYLE> w")
	if got != "x < y z w" {
		t.Errorf("PlainText fallback = %q", got)
	}
	if got, _ = PlainText("x < y <SCRIPT type=\"a\">\nalert(1)\n</SCRIPT> z"); got != "x < y z" {
		t.Errorf("PlainText fallback, upper-case multi-line script: %q", got)
	}
	got, _ = HTMLAbstract("x < y <script>alert(1)</script> z")
	if got != "x &lt; y z" {
		t.Errorf("HTMLAbstract fallback = %q", got)
	}
}

// formulaOf wraps publisher TeX in the JATS a record would carry it in.
func formulaOf(tex string) string {
	return "a <jats:inline-formula><jats:tex-math>" + tex + "</jats:tex-math></jats:inline-formula> b"
}

// TestTeXFallback_Unsafe checks publisher TeX used for a formula that has no
// MathML is dropped, with a warning, unless every command in it is on the safe
// list: no link, macro, include, class, style or delimiter-breaking command.
func TestTeXFallback_Unsafe(t *testing.T) {
	for _, tex := range []string{
		`\href{javascript:alert(1)}{x}`,
		`$\href{javascript:alert(1)}{x}$`,
		`\url{http://evil.example}`,
		`\require{html}`,
		`\class{a}{x}`,
		`\style{color:red}{x}`,
		`\cssId{a}{x}`,
		`\unicode{x41}`,
		`\def\a{b}`,
		`\newcommand{\a}{b}`,
		`\let\a\b`,
		`\input{x}`,
		`\char"41`,
		`\begin{matrix}a\end{matrix}`,
		`\mathchoice{a}{b}{c}{d}`,
		`x\) &lt;img src=x onerror=alert(1)&gt; \(y`,
		`x\)y`,
		`x\]`,
		`\[x`,
		`x\[`,
		`\(`,
		`x\(y`,
		`\[x\]\[y\]`,
		"x % comment \\)",
		`a % b`,
		`a$b`,
		`x\`,
		`\x{}`,
		`\ALPHA`,
		`\mathrm{a}\href{x}{y}`,
		`\text{\href{x}{y}}`,
	} {
		got, ws := HTMLAbstract(formulaOf(tex))
		if got != "a b" {
			t.Errorf("unsafe TeX %q was kept: %q", tex, got)
		}
		if len(ws) != 1 || ws[0].Code != WarnMathUnconverted || ws[0].Fragment != "tex-math" {
			t.Errorf("unsafe TeX %q: warnings %+v", tex, ws)
		}
	}
}

// TestTeXFallback_Safe checks ordinary publisher TeX is still used.
func TestTeXFallback_Safe(t *testing.T) {
	for _, tex := range []string{
		`x^2`, `x^{2}`, `x_{i}^{2}`, `\frac{a}{b}`, `\alpha+\beta`, `\sqrt{x}`, `\sqrt[3]{x}`,
		`a\,b`, `a\;b`, `\mathrm{d}x`, `\left(x\right)`, `\text{hi}`, `90\%`, `\sum_{i=0}^{n}`,
		`\mathbb{Z}`, `\overline{D}`, `\Lambda_{c}^{+}`, `J/\psi`, `a\_b`, `\{a\}`, `a\&b`, `a\#b`,
		`a\$b`, `a\:b`, `a\!b`, `a\ b`, `a\|b`, `a\\b`,
		`\sim 2`, `n - 1`, `\int_0^1 f(x)\,dx`, `\operatorname{sinc}x`, `\mathcal{L}`, `\hbar\omega`,
	} {
		got, ws := HTMLAbstract(formulaOf(tex))
		got = strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">").Replace(got)
		want := `a \(` + tex + `\) b`
		if got != want || len(ws) != 0 {
			t.Errorf("safe TeX %q: got %q, warnings %v", tex, got, ws)
		}
	}
	// layout newlines inside publisher TeX do not reach the output
	if got, _ := HTMLAbstract(formulaOf("\n  x^2\n ")); got != `a \(x^2\) b` {
		t.Errorf("whitespace: %q", got)
	}
	if got, _ := HTMLAbstract(formulaOf("a\n  +\n  b")); got != `a \(a + b\) b` {
		t.Errorf("inner whitespace: %q", got)
	}
}

// TestTeXFallback_TableCommands checks the safe list includes the commands
// the symbol table itself produces, from each of its three sections, so an
// override can add its own. The command names are made up so that the fixed
// list cannot be what lets them through.
func TestTeXFallback_TableCommands(t *testing.T) {
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.TeX["ℵ"] = `\zzsymbol`
	sym.TeX["ℶ"] = `^{\zzcircle}`
	sym.Accents["^"] = `\zzaccent`
	sym.MathVariants["blackboard"] = `\zzfont`
	for _, tex := range []string{`\zzsymbol`, `\zzcircle`, `\zzaccent{x}`, `\zzfont{1}`, `\tau`, `\Lambda`, `\to`} {
		if got, _ := htmlAbstract(formulaOf(tex), sym); got != `a \(`+tex+`\) b` {
			t.Errorf("%q with the override: %q", tex, got)
		}
	}
	// without the override the same made-up commands are refused
	for _, tex := range []string{`\zzsymbol`, `\zzcircle`, `\zzaccent{x}`, `\zzfont{1}`} {
		if got, _ := HTMLAbstract(formulaOf(tex)); got != "a b" {
			t.Errorf("%q without the override: %q", tex, got)
		}
	}
	// the fixed list covers what the default table does not
	if got, _ := HTMLAbstract(formulaOf(`\aleph`)); got != `a \(\aleph\) b` {
		t.Errorf("\\aleph is on the fixed list: %q", got)
	}
}

var (
	outTagRE   = regexp.MustCompile(`<(/?)([a-z0-9]+)((?: [a-z]+="[^"<>]*")*)>`)
	outAttrRE  = regexp.MustCompile(` ([a-z]+)="([^"]*)"`)
	allowedOut = map[string]bool{"href": true, "title": true, "colspan": true, "rowspan": true}
)

// checkSafeHTML reports why html is not safe to put in a description, or "".
// It is stricter than a sanitizer: every "<" must begin a tag HTMLAbstract writes,
// from the sanitizer's list, with only the attributes it keeps.
func checkSafeHTML(html string) string {
	for _, m := range outTagRE.FindAllStringSubmatch(html, -1) {
		if !sanitizerTags[m[2]] {
			return "tag " + m[2]
		}
		for _, a := range outAttrRE.FindAllStringSubmatch(m[3], -1) {
			if !allowedOut[a[1]] {
				return "attribute " + a[1]
			}
			if a[1] == "href" && !safeURL(strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`).Replace(a[2])) {
				return "href " + a[2]
			}
		}
	}
	rest := outTagRE.ReplaceAllString(html, "")
	if strings.ContainsAny(rest, "<>") {
		return "a stray angle bracket"
	}
	if strings.Contains(rest, "&") {
		for _, bit := range strings.Split(rest, "&")[1:] {
			if !strings.HasPrefix(bit, "amp;") && !strings.HasPrefix(bit, "lt;") && !strings.HasPrefix(bit, "gt;") {
				return "a bare ampersand: &" + bit[:min(len(bit), 10)]
			}
		}
	}
	return ""
}

// attackSeeds are strings that have broken or could break an HTML converter.
var attackSeeds = []string{
	``, `<`, `>`, `&`, `<a`, `<a href`, `<a href="`, `"><script>alert(1)</script>`,
	`<jats:p onclick="x()">a</jats:p>`, `<a href="javascript:alert(1)">x</a>`,
	`<a href="&#106;avascript:alert(1)">x</a>`, `<a href="java&#9;script:alert(1)">x</a>`,
	`<a href=" JAVASCRIPT:alert(1)">x</a>`, `<a href="data:text/html,<script>1</script>">x</a>`,
	`<img src=x onerror=alert(1)>`, `<svg onload=alert(1)>`, `<iframe src="javascript:1"></iframe>`,
	`<jats:inline-formula><jats:tex-math>\href{javascript:1}{x}</jats:tex-math></jats:inline-formula>`,
	`<jats:p>\(\href{javascript:1}{x}\)</jats:p>`, `<mml:math><mml:mtext>\href{x}{y}</mml:mtext></mml:math>`,
	`&lt;script&gt;alert(1)&lt;/script&gt;`, `&#0;`, `&#x110000;`, `&amp;amp;`, "a\x00b", "a\x01b\x7fc",
	`<![CDATA[<script>1</script>]]>`, `<!-- <script>1</script> -->`, `<?php echo 1 ?>`,
	`<!DOCTYPE x [<!ENTITY a "aaaaaaaaaa"><!ENTITY b "&a;&a;&a;&a;">]>&b;`,
	`<jats:title>Abstract</jats:title><jats:title>Abstract</jats:title>`,
	`<math><mi mathvariant="\href">x</mi></math>`, `<td colspan="999999999999999999999">x</td>`,
	`<p><p><p></p></p>`, `</p></p>`, `<br/><br>`, `<a href="/x" title="&quot;onmouseover=&quot;alert(1)">x</a>`,
}

// FuzzHTMLAbstract checks HTMLAbstract never panics and always writes safe HTML.
// With go test it runs over the seeds and every fixture; run it for longer with
// go test -fuzz=FuzzHTMLAbstract -fuzztime=60s.
func FuzzHTMLAbstract(f *testing.F) {
	for _, s := range attackSeeds {
		f.Add(s)
	}
	f.Add(`<jats:p>Water is H<jats:sub>2</jats:sub>O, <mml:math xmlns:mml="` + mathNS + `"><mml:mi>x</mml:mi></mml:math>.</jats:p>`)
	f.Fuzz(func(t *testing.T, s string) {
		got, _ := HTMLAbstract(s)
		if why := checkSafeHTML(got); why != "" {
			t.Fatalf("unsafe output (%s) for %q:\n%s", why, s, got)
		}
		for _, r := range got {
			if r < 0x20 || r == 0xFFFE || r == 0xFFFF {
				t.Fatalf("control or invalid character %U in output for %q", r, s)
			}
		}
	})
}

// FuzzPlainText checks PlainText never panics, never leaves markup it should
// have removed, and leaves no layout whitespace.
func FuzzPlainText(f *testing.F) {
	for _, s := range attackSeeds {
		f.Add(s)
	}
	f.Add("<i>Euclid</i>\n   preparation")
	f.Fuzz(func(t *testing.T, s string) {
		got, _ := PlainText(s)
		if got != strings.TrimSpace(got) || strings.ContainsAny(got, "\n\r\t") {
			t.Fatalf("untidy output %q for %q", got, s)
		}
		for _, r := range got {
			if r < 0x20 || r == 0xFFFE || r == 0xFFFF {
				t.Fatalf("control or invalid character %U in output for %q", r, s)
			}
		}
	})
}

// TestFuzzSeedsFromFixtures runs the same checks over every real title and
// abstract, so the invariants hold on real data and not only on the seeds.
func TestFuzzSeedsFromFixtures(t *testing.T) {
	for name, abs := range fixtureAbstracts(t) {
		got, _ := HTMLAbstract(abs)
		if why := checkSafeHTML(got); why != "" {
			t.Errorf("%s: unsafe output (%s)", name, why)
		}
	}
}

// TestRobust_WarningFragmentLength checks a parse-fallback warning does not
// carry a huge input, and never cuts a character in half.
func TestRobust_WarningFragmentLength(t *testing.T) {
	short := "a < b"
	if _, ws := PlainText(short); len(ws) != 1 || ws[0].Fragment != short {
		t.Errorf("a short input is quoted whole: %+v", ws)
	}
	long := "< " + strings.Repeat("é", 2000)
	_, ws := PlainText(long)
	if len(ws) != 1 {
		t.Fatalf("warnings %+v", ws)
	}
	f := ws[0].Fragment
	if len(f) > maxFragment+len("…") || !strings.HasSuffix(f, "…") || !strings.HasPrefix(f, "< é") {
		t.Errorf("fragment is %d bytes: %.30q…%q", len(f), f, f[max(0, len(f)-6):])
	}
	if !strings.HasPrefix(long, strings.TrimSuffix(f, "…")) {
		t.Error("the shortened fragment is not a prefix of the input")
	}
	edge := strings.Repeat("a", maxFragment) + "é"
	if got := shorten(edge); got != strings.Repeat("a", maxFragment)+"…" {
		t.Errorf("a cut at the limit: %d bytes", len(got))
	}
	cut := strings.Repeat("a", maxFragment-1) + "éé"
	if got := shorten(cut); got != strings.Repeat("a", maxFragment-1)+"…" {
		t.Errorf("a cut inside a character must back up: %q", got[len(got)-6:])
	}
	if exact := strings.Repeat("a", maxFragment); shorten(exact) != exact {
		t.Error("an input at the limit is quoted whole")
	}
}

// TestRobust_InvalidXMLCharacters checks characters XML does not allow never
// reach the output, whichever way they arrive, and the allowed ones stay.
func TestRobust_InvalidXMLCharacters(t *testing.T) {
	cases := []struct{ name, in, plain string }{
		{"a control character before a space, through the fallback", "\x00 <", "<"},
		{"control characters between words, through the fallback", "a \x01 \x02 < b", "a < b"},
		{"numeric reference", "a&#1;b", "ab"},
		{"numeric reference to zero", "a&#0;b", "ab"},
		{"numeric reference past the last character", "a&#x110000;b", "ab"},
		{"numeric reference to a surrogate", "a&#xD800;b", "ab"},
		{"numeric reference to a huge number", "a&#99999999999999999999;b", "ab"},
		{"numeric reference to a non-character", "a&#xFFFE;b", "ab"},
		{"numeric reference to tab is kept as space", "a&#9;b", "a b"},
		{"valid numeric reference", "caf&#233; &#x41;", "café A"},
		{"the first character above the controls", "a&#32;b", "a b"},
		{"numeric reference, hex", "a&#x7;b", "ab"},
		{"raw control character", "a\x01b", "ab"},
		{"through the fallback", "a < \x01b￾c￿d", "a < bcd"},
		{"through a fallback entity", "a < b&#2;c", "a < bc"},
		{"tab and newline are whitespace, not dropped", "a\tb\nc", "a b c"},
		{"characters beyond the BMP stay", "a\U0001F600b \U0010FFFF", "a\U0001F600b \U0010FFFF"},
		{"the last allowed BMP character", "a�b", "a�b"},
		{"the first allowed character after the controls", "a b", "a b"},
		{"the start of the private-use range", "ab", "ab"},
		{"the end of the range before the surrogates", "a퟿b", "a퟿b"},
	}
	for _, c := range cases {
		if got, _ := PlainText(c.in); got != c.plain {
			t.Errorf("%s: PlainText(%q) = %q, want %q", c.name, c.in, got, c.plain)
		}
	}
	// a stray character is removed without costing the abstract its structure
	if got, ws := HTMLAbstract("<jats:p>a&#1;b</jats:p><jats:p>c</jats:p>"); got != "<p>ab</p><p>c</p>" || len(ws) != 0 {
		t.Errorf("HTMLAbstract text: %q, %v", got, ws)
	}
	if got, ws := HTMLAbstract("<jats:p>a\x01b</jats:p><jats:p>c</jats:p>"); got != "<p>ab</p><p>c</p>" || len(ws) != 0 {
		t.Errorf("a raw control character keeps the structure: %q, %v", got, ws)
	}
	if got, ws := HTMLAbstract(`<a href="/x&#1;y" title="t&#2;u">l</a>`); got != `<a href="/xy" title="tu">l</a>` || len(ws) != 0 {
		t.Errorf("HTMLAbstract attributes: %q, %v", got, ws)
	}
	if got, ws := PlainText("a&#1;b <i>c</i>"); got != "ab c" || len(ws) != 0 {
		t.Errorf("PlainText keeps converting: %q, %v", got, ws)
	}
}

// TestRobust_ItalicGlueUnchanged checks the cheaper one-letter-italic test
// agrees with the old one on the cases that matter.
func TestRobust_ItalicGlueUnchanged(t *testing.T) {
	for in, want := range map[string]string{
		"a\n <i>x</i>\n b":                     "axb",
		"a\n <i> x </i>\n b":                   "axb",
		"a\n <i>xy</i>\n b":                    "a xy b",
		"a\n <i><b>x</b></i>\n b":              "axb",
		"a\n <i><sub>2</sub></i>\n b":          "a₂b",
		"a\n <i>x<b></b></i>\n b":              "axb",
		"a\n <i></i>\n b":                      "a b",
		"a\n <jats:italic>é</jats:italic>\n b": "aéb",
	} {
		if got, _ := PlainText(in); got != want {
			t.Errorf("PlainText(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestRobust_NestingBoundary pins the limit: 254 nested elements convert and
// 255 fall back (the parser's own root and wrapper count as two levels).
func TestRobust_NestingBoundary(t *testing.T) {
	nest := func(d int) string { return strings.Repeat("<span>", d) + "x" + strings.Repeat("</span>", d) }
	if got, ws := HTMLAbstract(nest(maxNesting - 2)); len(ws) != 0 || strings.Count(got, "<span>") != maxNesting-2 {
		t.Errorf("%d nested: %d spans, warnings %v", maxNesting-2, strings.Count(got, "<span>"), ws)
	}
	if got, ws := HTMLAbstract(nest(maxNesting - 1)); got != "x" || len(ws) != 1 {
		t.Errorf("%d nested: %.30q, warnings %v", maxNesting-1, got, ws)
	}
}

// TestRobust_MoreInvalidCharacters pins both ends of each allowed range.
func TestRobust_MoreInvalidCharacters(t *testing.T) {
	for in, want := range map[string]string{
		"a\x1fb":       "ab",
		"a\x0bb":       "ab",
		"a\x0cb":       "ab",
		"a\rb":         "a b",
		"a&#xDFFF;b":   "ab",
		"a&#xE000;b":   "ab",
		"a\U00010000b": "a\U00010000b",
		"a&#x10000;b":  "a\U00010000b",
		"a&#x10FFFF;b": "a\U0010FFFFb",
		"a&#xFFFD;b":   "a�b",
		"a&#X41;b":     "a&#X41;b",
		"a&#x;b":       "a&#x;b",
	} {
		if got, _ := PlainText(in); got != want {
			t.Errorf("PlainText(%q) = %q, want %q", in, got, want)
		}
	}
}
