package crossrefapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// mathNodes returns the MathML formula nodes of a parsed fragment, in order.
func mathNodes(nodes []*node) []*node {
	var out []*node
	for _, n := range nodes {
		if n.name != "" && (n.space == mathNS || n.name == "math") {
			out = append(out, n)
			continue
		}
		out = append(out, mathNodes(n.kids)...)
	}
	return out
}

// fixtureAbstract returns the abstract of a captured CrossRef response.
func fixtureAbstract(t *testing.T, doiFile string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "crossref-"+doiFile+".json"))
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
	return r.Message.Abstract
}

// texOfFragment parses a fragment and returns the TeX of each formula in it.
func texOfFragment(t *testing.T, src string) ([]string, []Warning) {
	t.Helper()
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	root, err := parseFragment(src, sym)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out []string
	var ws []Warning
	for _, m := range mathNodes(root.kids) {
		out = append(out, mathTeX(m, sym, &ws))
	}
	return out, ws
}

// TestMathTeX_Fixtures checks every formula of real captured abstracts. The
// expected TeX is written from the conversion rules (and, for the Springer
// formulas, agrees with the publisher's own TeX; see the oracle test).
func TestMathTeX_Fixtures(t *testing.T) {
	cases := []struct {
		file string
		want []string
	}{
		{"10.1007_s41114-018-0012-9", []string{`90\%`, `20~\mathrm{deg}^{2}`, `\sim 2`}},
		{"10.1007_s00332-025-10238-4", []string{`n-1`}},
		{"10.1103_physrevb.109.035167", []string{`1T-\mathrm{TaS}_{2}`, `1T-\mathrm{TaS}_{2}`, `d_{z^{2}}`}},
		{"10.1103_physrevlett.134.141902", []string{
			`\mathbb{Z}_{N}`, `\mathrm{SU}(N)`, `\mathbb{Z}_{N}`, `d=2`, `d=3`, `d=4`, `\mathbb{Z}_{N}`, `N-1`,
		}},
		{"10.1103_sh13-grfw", []string{
			`J/\psi`,
			`\overline{D}^{0}\mathrm{\Lambda}_{c}^{+}`,
			`\overline{D}^{*0}\mathrm{\Lambda}_{c}^{+}`,
			`t`, `s`, `u`, `t`,
			`J/\psi-007`,
			`\overline{D}^{0}\mathrm{\Lambda}_{c}^{+}`,
			`\overline{D}^{*0}\mathrm{\Lambda}_{c}^{+}`,
			`\gamma p\to\overline{D}^{(*)0}\mathrm{\Lambda}_{c}^{+}`,
		}},
		// a prescript: an empty base needs {} so the scripts attach to nothing
		{"10.3847_1538-4357_ae99c4", []string{`{}_{4}^{+}`}},
	}
	for _, c := range cases {
		sym, _ := DefaultSymbols()
		root, err := parseFragment(fixtureAbstract(t, c.file), sym)
		if err != nil {
			t.Fatalf("%s: %v", c.file, err)
		}
		var got []string
		var ws []Warning
		for _, m := range mathNodes(root.kids) {
			got = append(got, mathTeX(m, sym, &ws))
		}
		if len(got) < len(c.want) {
			t.Errorf("%s: %d formulas, want at least %d", c.file, len(got), len(c.want))
			continue
		}
		for i, w := range c.want {
			if got[i] != w {
				t.Errorf("%s formula %d:\n got  %s\n want %s", c.file, i, got[i], w)
			}
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.file, ws)
		}
	}
}

// TestMathTeX_PublisherOracle compares our TeX with the TeX the publisher
// (Springer) supplied beside the MathML in <jats:tex-math>, after removing
// whitespace and the braces around one-character groups. Decision E: the
// publisher's TeX is never used in the output, only as a check on ours.
func TestMathTeX_PublisherOracle(t *testing.T) {
	norm := func(s string) string {
		s = strings.Join(strings.Fields(s), "")
		s = strings.TrimSuffix(strings.TrimPrefix(s, "$$"), "$$")
		return regexp.MustCompile(`([_^])\{([^{}\\])\}`).ReplaceAllString(s, "$1$2")
	}
	texRE := regexp.MustCompile(`(?s)<jats:tex-math>(.*?)</jats:tex-math>`)
	pairs := 0
	for _, file := range []string{"10.1007_s41114-018-0012-9", "10.1007_s00332-025-10238-4"} {
		abs := fixtureAbstract(t, file)
		theirs := texRE.FindAllStringSubmatch(abs, -1)
		ours, _ := texOfFragment(t, abs)
		if len(theirs) == 0 || len(theirs) != len(ours) {
			t.Fatalf("%s: %d publisher TeX strings and %d formulas", file, len(theirs), len(ours))
		}
		for i := range ours {
			pairs++
			if norm(ours[i]) != norm(theirs[i][1]) {
				t.Errorf("%s formula %d: ours %q, publisher %q", file, i, norm(ours[i]), norm(theirs[i][1]))
			}
		}
	}
	if pairs != 4 {
		t.Errorf("checked %d pairs, want 4", pairs)
	}
}

// TestMathTeX_Elements covers each element the plan names, with SYNTHETIC
// MathML written from the specification: the real abstracts contain no mfrac,
// msqrt, mroot, munder, munderover or mtext.
func TestMathTeX_Elements(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"mn mi mo", `<mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>12.5</mml:mn>`, `x+12.5`},
		{"greek letter", `<mml:mi>α</mml:mi>`, `\alpha`},
		{"command then letter gets a space", `<mml:mi>α</mml:mi><mml:mi>x</mml:mi>`, `\alpha x`},
		{"command then digit gets a space", `<mml:mo>∼</mml:mo><mml:mn>2</mml:mn>`, `\sim 2`},
		{"command then bracket gets none", `<mml:mi>α</mml:mi><mml:mo>(</mml:mo>`, `\alpha(`},
		{"two commands", `<mml:mi>α</mml:mi><mml:mi>β</mml:mi>`, `\alpha\beta`},
		{"multi-letter identifier is upright", `<mml:mi>sin</mml:mi><mml:mi>x</mml:mi>`, `\mathrm{sin}x`},
		{"operators from the table", `<mml:mi>a</mml:mi><mml:mo>×</mml:mo><mml:mi>b</mml:mi><mml:mo>±</mml:mo><mml:mi>c</mml:mi>`, `a\times b\pm c`},
		{"minus sign", `<mml:mi>a</mml:mi><mml:mo>−</mml:mo><mml:mi>b</mml:mi>`, `a-b`},
		{"mrow and mstyle are transparent", `<mml:mstyle><mml:mrow><mml:mi>a</mml:mi><mml:mrow><mml:mo>=</mml:mo><mml:mi>b</mml:mi></mml:mrow></mml:mrow></mml:mstyle>`, `a=b`},
		{"mtext", `<mml:mtext>for all</mml:mtext>`, `\text{for all}`},
		{"mtext of one symbol uses the table", `<mml:mtext>−</mml:mtext>`, `-`},
		{"msup", `<mml:msup><mml:mi>x</mml:mi><mml:mn>2</mml:mn></mml:msup>`, `x^{2}`},
		{"msub", `<mml:msub><mml:mi>x</mml:mi><mml:mn>1</mml:mn></mml:msub>`, `x_{1}`},
		{"msup with a long base is grouped", `<mml:msup><mml:mrow><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:mrow><mml:mn>2</mml:mn></mml:msup>`, `{ab}^{2}`},
		{"msubsup", `<mml:msubsup><mml:mi>x</mml:mi><mml:mi>i</mml:mi><mml:mn>2</mml:mn></mml:msubsup>`, `x_{i}^{2}`},
		{"mfrac", `<mml:mfrac><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:mfrac>`, `\frac{a}{b}`},
		{"mfrac of expressions", `<mml:mfrac><mml:mrow><mml:mi>a</mml:mi><mml:mo>+</mml:mo><mml:mi>b</mml:mi></mml:mrow><mml:mn>2</mml:mn></mml:mfrac>`, `\frac{a+b}{2}`},
		{"msqrt", `<mml:msqrt><mml:mi>x</mml:mi></mml:msqrt>`, `\sqrt{x}`},
		{"mroot", `<mml:mroot><mml:mi>x</mml:mi><mml:mn>3</mml:mn></mml:mroot>`, `\sqrt[3]{x}`},
		{"munderover sum", `<mml:munderover><mml:mo>∑</mml:mo><mml:mrow><mml:mi>i</mml:mi><mml:mo>=</mml:mo><mml:mn>0</mml:mn></mml:mrow><mml:mi>n</mml:mi></mml:munderover><mml:msub><mml:mi>a</mml:mi><mml:mi>i</mml:mi></mml:msub>`, `\sum_{i=0}^{n}a_{i}`},
		{"munder", `<mml:munder><mml:mi>lim</mml:mi><mml:mrow><mml:mi>x</mml:mi><mml:mo>→</mml:mo><mml:mn>0</mml:mn></mml:mrow></mml:munder>`, `\mathrm{lim}_{x\to 0}`},
		{"mover accent", `<mml:mover accent="true"><mml:mi>D</mml:mi><mml:mo>¯</mml:mo></mml:mover>`, `\overline{D}`},
		{"mover hat", `<mml:mover><mml:mi>x</mml:mi><mml:mo>^</mml:mo></mml:mover>`, `\hat{x}`},
		{"mover not an accent", `<mml:mover><mml:mi>x</mml:mi><mml:mi>y</mml:mi></mml:mover>`, `x^{y}`},
		{"accent inside an mrow", `<mml:mover><mml:mi>x</mml:mi><mml:mrow><mml:mo>^</mml:mo></mml:mrow></mml:mover>`, `\hat{x}`},
		{"specials in a number", `<mml:mn>1_000</mml:mn>`, `1\_000`},
		{"empty token adds nothing", `<mml:mi>a</mml:mi><mml:mi/><mml:mi>b</mml:mi>`, `ab`},
		{"empty row", `<mml:mrow/>`, ``},
		{"nested script as a base is grouped", `<mml:msup><mml:msup><mml:mi>x</mml:mi><mml:mn>2</mml:mn></mml:msup><mml:mn>3</mml:mn></mml:msup>`, `{x^{2}}^{3}`},
		{"empty number is not a token", `<mml:msup><mml:mrow><mml:mi>a</mml:mi><mml:mn/></mml:mrow><mml:mn>2</mml:mn></mml:msup>`, `a^{2}`},
		{"text directly inside an mrow", `<mml:mrow>x<mml:mi>y</mml:mi></mml:mrow>`, `xy`},
		{"specials in a multi-letter identifier", `<mml:mi>a_b</mml:mi>`, `\mathrm{a\_b}`},
		{"special in a one-letter identifier", `<mml:mi>#</mml:mi>`, `\#`},
		{"mathvariant from the table", `<mml:mi mathvariant="double-struck">R</mml:mi>`, `\mathbb{R}`},
		{"mathvariant normal on a letter", `<mml:mi mathvariant="normal">d</mml:mi>`, `\mathrm{d}`},
		{"mathvariant on a multi-letter identifier", `<mml:mi mathvariant="bold">TaS</mml:mi>`, `\mathbf{TaS}`},
		{"unknown mathvariant is ignored", `<mml:mi mathvariant="no-such-variant">x</mml:mi>`, `x`},
		{"mspace", `<mml:mn>1</mml:mn><mml:mspace/><mml:mn>2</mml:mn>`, `1~2`},
		{"invisible operators removed", "<mml:mi>a</mml:mi><mml:mo>\u2062</mml:mo><mml:mi>b</mml:mi>", `ab`},
		// TeX special characters in text are escaped
		{"specials in mtext", `<mml:mtext>50% &amp; #1_a {b} $c</mml:mtext>`, `\text{50\% \& \#1\_a \{b\} \$c}`},
		{"specials in an operator", `<mml:mo>#</mml:mo><mml:mo>&amp;</mml:mo><mml:mo>_</mml:mo><mml:mo>{</mml:mo><mml:mo>}</mml:mo>`, `\#\&\_\{\}`},
		{"percent in an operator", `<mml:mo>%</mml:mo>`, `\%`},
		{"backslash in mtext", `<mml:mtext>a\b</mml:mtext>`, `\text{a\textbackslash{}b}`},
		{"layout whitespace", "\n <mml:mi>a</mml:mi>\n <mml:mo>=</mml:mo>\n <mml:mi>b</mml:mi>\n", `a=b`},
		{"empty", ``, ``},
	}
	for _, c := range cases {
		got, ws := texOfFragment(t, mml(c.body))
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.name, ws)
		}
	}
}

// TestMathTeX_Unconverted checks elements with no conversion: their text is
// kept inside \text{} and a math-unconverted warning names the element.
func TestMathTeX_Unconverted(t *testing.T) {
	cases := []struct{ name, body, want, el string }{
		{"mtable", `<mml:mtable><mml:mtr><mml:mtd><mml:mi>a</mml:mi></mml:mtd><mml:mtd><mml:mi>b</mml:mi></mml:mtd></mml:mtr></mml:mtable>`, `\text{ab}`, "mtable"},
		{"semantics", `<mml:semantics><mml:mi>x</mml:mi><mml:annotation>TeX</mml:annotation></mml:semantics>`, `\text{xTeX}`, "semantics"},
		{"special in unconverted text", `<mml:menclose><mml:mn>5%</mml:mn></mml:menclose>`, `\text{5\%}`, "menclose"},
		{"msup with one child", `<mml:msup><mml:mi>x</mml:mi></mml:msup>`, `\text{x}`, "msup"},
		{"mfrac with three children", `<mml:mfrac><mml:mi>a</mml:mi><mml:mi>b</mml:mi><mml:mi>c</mml:mi></mml:mfrac>`, `\text{abc}`, "mfrac"},
		{"mover with one child", `<mml:mover><mml:mi>a</mml:mi></mml:mover>`, `\text{a}`, "mover"},
		{"msup with three children", `<mml:msup><mml:mi>a</mml:mi><mml:mi>b</mml:mi><mml:mi>c</mml:mi></mml:msup>`, `\text{abc}`, "msup"},
		{"munder with one child", `<mml:munder><mml:mi>a</mml:mi></mml:munder>`, `\text{a}`, "munder"},
		{"munderover with two children", `<mml:munderover><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:munderover>`, `\text{ab}`, "munderover"},
		{"mroot with one child", `<mml:mroot><mml:mi>a</mml:mi></mml:mroot>`, `\text{a}`, "mroot"},
		{"empty unconverted element", `<mml:mtable/>`, ``, "mtable"},
		{"msubsup with four children", `<mml:msubsup><mml:mi>a</mml:mi><mml:mi>b</mml:mi><mml:mi>c</mml:mi><mml:mi>d</mml:mi></mml:msubsup>`, `\text{abcd}`, "msubsup"},
		{"mover with three children", `<mml:mover><mml:mi>a</mml:mi><mml:mi>b</mml:mi><mml:mi>c</mml:mi></mml:mover>`, `\text{abc}`, "mover"},
		{"mroot with three children", `<mml:mroot><mml:mi>a</mml:mi><mml:mi>b</mml:mi><mml:mi>c</mml:mi></mml:mroot>`, `\text{abc}`, "mroot"},
		{"msubsup with two children", `<mml:msubsup><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:msubsup>`, `\text{ab}`, "msubsup"},
	}
	for _, c := range cases {
		got, ws := texOfFragment(t, mml(c.body))
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
		if len(ws) != 1 || ws[0].Code != WarnMathUnconverted || ws[0].Fragment != c.el {
			t.Errorf("%s: warnings %+v, want one %s for %q", c.name, ws, WarnMathUnconverted, c.el)
		}
	}
}

// TestMathTeX_TableOverride checks the table drives the output.
func TestMathTeX_TableOverride(t *testing.T) {
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.TeX["×"] = `\cdot`
	sym.Accents["^"] = `\widehat`
	sym.MathVariants["bold"] = `\boldsymbol`
	root, err := parseFragment(mml(`<mml:mi mathvariant="bold">a</mml:mi><mml:mo>×</mml:mo><mml:mover><mml:mi>x</mml:mi><mml:mo>^</mml:mo></mml:mover>`), sym)
	if err != nil {
		t.Fatal(err)
	}
	var ws []Warning
	got := mathTeX(mathNodes(root.kids)[0], sym, &ws)
	if got != `\boldsymbol{a}\cdot\widehat{x}` || len(ws) != 0 {
		t.Errorf("got %q, warnings %v", got, ws)
	}
}
