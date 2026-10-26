package crossrefapi

import (
	"strings"
	"testing"
)

// mml wraps MathML body text in a <math> element with the "mml" prefix.
func mml(body string) string {
	return `<mml:math xmlns:mml="` + mathNS + `">` + body + `</mml:math>`
}

// TestMathML_Fixtures checks real captured titles end to end, formulas
// included. Expected strings come from the plan and the decisions, not from
// the code's output.
func TestMathML_Fixtures(t *testing.T) {
	cases := []struct {
		file string
		want string
		warn []string
	}{
		// msup: 1/r² with the "/" an operator, and a space where the formula abuts words
		{"10.1103_physrevb.85.144303", "Line of continuous phase transitions in a three-dimensional U(1) loop model with 1/r² current-current interactions", nil},
		// msub with a mapped subscript, an mtext minus
		{"10.1103_physrevb.109.035167", "Gap collapse and flat band induced by uniaxial strain in 1T−TaS₂", nil},
		// mi and mo tokens, layout whitespace inside the formula
		{"10.1103_sh13-grfw", "Rescattering effects in near-threshold J/ψ photoproduction", nil},
		// a subscript the table cannot show: linear form, with a warning
		{"10.1103_physrevlett.134.141902", "Phases of Theories with Z_N 1-Form Symmetry, and the Roles of Center Vortices and Magnetic Monopoles", []string{WarnMathLinear}},
		// a formula first in the title, then a hyphen
		{"10.1103_physrevd.111.013006", "τ-lepton pair spin in proton-proton LHC collisions for anomalous dipole moments", nil},
	}
	for _, c := range cases {
		title := fixtureTitles(t, c.file)[0]
		got, ws := PlainText(title)
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.file, got, c.want)
		}
		if strings.Join(codes(ws), ",") != strings.Join(c.warn, ",") {
			t.Errorf("%s: warnings %v, want %v", c.file, codes(ws), c.warn)
		}
	}
}

// TestMathML_Elements covers each element the plan names. The real titles in
// both surveys use only msup, msub, mrow, mi, mn, mo and mtext, so the fraction,
// root and unrecognised-element cases below are SYNTHETIC: they are written
// from the MathML specification, not captured from CrossRef.
func TestMathML_Elements(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"token elements", `<mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>12</mml:mn>`, "x+12"},
		{"mrow concatenates", `<mml:mrow><mml:mi>a</mml:mi><mml:mrow><mml:mo>=</mml:mo><mml:mi>b</mml:mi></mml:mrow></mml:mrow>`, "a=b"},
		{"mstyle and mpadded are transparent", `<mml:mstyle><mml:mpadded><mml:mi>x</mml:mi></mml:mpadded></mml:mstyle>`, "x"},
		{"mtext keeps its words", `<mml:mtext>for all</mml:mtext>`, "for all"},
		{"invisible operators are removed by the table", "<mml:mi>a</mml:mi><mml:mo>\u2062</mml:mo><mml:mi>b</mml:mi><mml:mo>\u2061</mml:mo>", "ab"},
		{"layout whitespace between elements", "\n <mml:mi>a</mml:mi>\n <mml:mo>=</mml:mo>\n <mml:mi>b</mml:mi>\n", "a=b"},
		// msup and msub
		{"msup maps to superscripts", `<mml:msup><mml:mi>x</mml:mi><mml:mn>2</mml:mn></mml:msup>`, "x²"},
		{"msup with a minus sign", `<mml:msup><mml:mn>10</mml:mn><mml:mrow><mml:mo>−</mml:mo><mml:mn>2</mml:mn></mml:mrow></mml:msup>`, "10⁻²"},
		{"msub maps to subscripts", `<mml:msub><mml:mi>x</mml:mi><mml:mn>1</mml:mn></mml:msub>`, "x₁"},
		{"msub with a letter subscript", `<mml:msub><mml:mi>x</mml:mi><mml:mi>i</mml:mi></mml:msub>`, "xᵢ"},
		// mfrac
		{"mfrac of single tokens", `<mml:mfrac><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:mfrac>`, "a/b"},
		{"mfrac with a long numerator", `<mml:mfrac><mml:mrow><mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>1</mml:mn></mml:mrow><mml:mn>2</mml:mn></mml:mfrac>`, "(x+1)/2"},
		{"mfrac with a long denominator", `<mml:mfrac><mml:mn>1</mml:mn><mml:mrow><mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>1</mml:mn></mml:mrow></mml:mfrac>`, "1/(x+1)"},
		{"mfrac with both long", `<mml:mfrac><mml:mrow><mml:mi>a</mml:mi><mml:mo>+</mml:mo><mml:mi>b</mml:mi></mml:mrow><mml:mrow><mml:mi>c</mml:mi><mml:mo>+</mml:mo><mml:mi>d</mml:mi></mml:mrow></mml:mfrac>`, "(a+b)/(c+d)"},
		// msqrt and mroot
		{"msqrt of one token", `<mml:msqrt><mml:mn>2</mml:mn></mml:msqrt>`, "√2"},
		{"msqrt of an expression", `<mml:msqrt><mml:mrow><mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>1</mml:mn></mml:mrow></mml:msqrt>`, "√(x+1)"},
		{"mroot with a mappable index", `<mml:mroot><mml:mi>x</mml:mi><mml:mn>3</mml:mn></mml:mroot>`, "³√x"},
		{"mroot of an expression", `<mml:mroot><mml:mrow><mml:mi>x</mml:mi><mml:mo>+</mml:mo><mml:mn>1</mml:mn></mml:mrow><mml:mn>3</mml:mn></mml:mroot>`, "³√(x+1)"},
	}
	for _, c := range cases {
		got, ws := PlainText("a " + mml(c.body) + " b")
		want := "a " + c.want + " b"
		if got != want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, want)
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.name, ws)
		}
	}
}

// TestMathML_LinearForms covers scripts and roots the Unicode forms cannot
// show: the linear form is written and a math-linear warning carries it.
func TestMathML_LinearForms(t *testing.T) {
	cases := []struct {
		name, body, want, fragment string
	}{
		{"unmappable superscript", `<mml:msup><mml:mi>x</mml:mi><mml:mi>q</mml:mi></mml:msup>`, "x^q", "x^q"},
		{"unmappable subscript", `<mml:msub><mml:mi>Z</mml:mi><mml:mi>N</mml:mi></mml:msub>`, "Z_N", "Z_N"},
		{"unmappable script of several tokens", `<mml:msup><mml:mi>x</mml:mi><mml:mrow><mml:mn>2</mml:mn><mml:mi>q</mml:mi></mml:mrow></mml:msup>`, "x^(2q)", "x^(2q)"},
		{"partly mappable script is all linear", `<mml:msub><mml:mi>x</mml:mi><mml:mrow><mml:mn>1</mml:mn><mml:mi>N</mml:mi></mml:mrow></mml:msub>`, "x_(1N)", "x_(1N)"},
		{"unmappable root index", `<mml:mroot><mml:mi>x</mml:mi><mml:mi>q</mml:mi></mml:mroot>`, "x^(1/q)", "x^(1/q)"},
		{"unmappable root index of an expression", `<mml:mroot><mml:mrow><mml:mi>a</mml:mi><mml:mo>+</mml:mo><mml:mi>b</mml:mi></mml:mrow><mml:mi>q</mml:mi></mml:mroot>`, "(a+b)^(1/q)", "(a+b)^(1/q)"},
	}
	for _, c := range cases {
		got, ws := PlainText("a " + mml(c.body) + " b")
		if got != "a "+c.want+" b" {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, "a "+c.want+" b")
		}
		if len(ws) != 1 || ws[0].Code != WarnMathLinear || ws[0].Fragment != c.fragment {
			t.Errorf("%s: warnings %+v, want one %s with fragment %q", c.name, ws, WarnMathLinear, c.fragment)
		}
	}
}

// TestMathML_Unconverted checks elements with no conversion: the text content
// is kept and a math-unconverted warning names the element. SYNTHETIC cases;
// none of these elements occurs in a real title in either survey (mover,
// msubsup and mspace occur in real abstracts, which Phase 4 handles).
func TestMathML_Unconverted(t *testing.T) {
	cases := []struct {
		name, body, want, element string
	}{
		{"mtable", `<mml:mtable><mml:mtr><mml:mtd><mml:mi>a</mml:mi></mml:mtd><mml:mtd><mml:mi>b</mml:mi></mml:mtd></mml:mtr></mml:mtable>`, "ab", "mtable"},
		{"munderover", `<mml:munderover><mml:mo>∑</mml:mo><mml:mn>0</mml:mn><mml:mi>n</mml:mi></mml:munderover>`, "∑0n", "munderover"},
		{"mover", `<mml:mover><mml:mi>x</mml:mi><mml:mo>¯</mml:mo></mml:mover>`, "x¯", "mover"},
		{"msubsup", `<mml:msubsup><mml:mi>x</mml:mi><mml:mn>1</mml:mn><mml:mn>2</mml:mn></mml:msubsup>`, "x12", "msubsup"},
		{"semantics", `<mml:semantics><mml:mi>x</mml:mi><mml:annotation>TeX</mml:annotation></mml:semantics>`, "xTeX", "semantics"},
	}
	for _, c := range cases {
		got, ws := PlainText("a " + mml(c.body) + " b")
		if got != "a "+c.want+" b" {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, "a "+c.want+" b")
		}
		found := false
		for _, w := range ws {
			if w.Code == WarnMathUnconverted && w.Fragment == c.element {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: want a %s warning naming %q, got %+v", c.name, WarnMathUnconverted, c.element, ws)
		}
	}
	// an unconverted element nested in a converted one still warns, once
	_, ws := PlainText("a " + mml(`<mml:msup><mml:mover><mml:mi>x</mml:mi><mml:mo>¯</mml:mo></mml:mover><mml:mn>2</mml:mn></mml:msup>`) + " b")
	if len(ws) != 1 || ws[0].Fragment != "mover" {
		t.Errorf("nested unconverted element: warnings %+v", ws)
	}
}

// TestMathML_TableOverride checks the symbol table drives the conversion: a
// token entry changes what a token becomes, and a superscript entry changes
// which scripts can be shown.
func TestMathML_TableOverride(t *testing.T) {
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.Tokens["×"] = "x"
	sym.Superscripts["q"] = "ᵠ"
	got, ws := plainText(mml(`<mml:mi>a</mml:mi><mml:mo>×</mml:mo><mml:msup><mml:mi>b</mml:mi><mml:mi>q</mml:mi></mml:msup>`), sym)
	if got != "axbᵠ" || len(ws) != 0 {
		t.Errorf("got %q, warnings %+v; want %q and none", got, ws, "axbᵠ")
	}
}

// TestMathML_EmptyFormula checks a formula with no content disappears.
func TestMathML_EmptyFormula(t *testing.T) {
	got, ws := PlainText("a " + mml("") + " b")
	if got != "a b" || len(ws) != 0 {
		t.Errorf("got %q, warnings %+v", got, ws)
	}
}

// TestMathML_Nesting checks the token count that decides brackets survives
// nesting: a script adds its tokens, and a fraction, root or square root counts
// as more than one.
func TestMathML_Nesting(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"script inside a fraction", `<mml:mfrac><mml:msup><mml:mi>x</mml:mi><mml:mn>2</mml:mn></mml:msup><mml:mn>3</mml:mn></mml:mfrac>`, "(x²)/3"},
		{"fraction inside a fraction", `<mml:mfrac><mml:mfrac><mml:mi>a</mml:mi><mml:mi>b</mml:mi></mml:mfrac><mml:mi>c</mml:mi></mml:mfrac>`, "(a/b)/c"},
		{"square root inside a square root", `<mml:msqrt><mml:msqrt><mml:mi>x</mml:mi></mml:msqrt></mml:msqrt>`, "√(√x)"},
		{"root inside a fraction", `<mml:mfrac><mml:mroot><mml:mi>x</mml:mi><mml:mn>3</mml:mn></mml:mroot><mml:mn>2</mml:mn></mml:mfrac>`, "(³√x)/2"},
		{"root with a long index", `<mml:mroot><mml:mi>x</mml:mi><mml:mrow><mml:mn>2</mml:mn><mml:mn>3</mml:mn></mml:mrow></mml:mroot>`, "²³√x"},
		{"empty token holds no token", `<mml:mfrac><mml:mrow><mml:mi>x</mml:mi><mml:mi></mml:mi></mml:mrow><mml:mn>2</mml:mn></mml:mfrac>`, "x/2"},
		{"text directly inside an mrow", `<mml:mrow>x<mml:mi>y</mml:mi></mml:mrow>`, "xy"},
		{"ms string literal", `<mml:ms>abc</mml:ms>`, "abc"},
		{"whitespace inside mtext collapses", "<mml:mtext>for \n   all</mml:mtext>", "for all"},
		{"table entry applies per character inside a token", "<mml:mi>x\u2062y</mml:mi>", "xy"},
	}
	for _, c := range cases {
		got, ws := PlainText("a " + mml(c.body) + " b")
		if got != "a "+c.want+" b" || len(ws) != 0 {
			t.Errorf("%s:\n got  %q (warnings %v)\n want %q", c.name, got, ws, "a "+c.want+" b")
		}
	}
}

// TestMathML_Malformed checks elements with the wrong number of children are
// reduced to their text, with a warning naming the element.
func TestMathML_Malformed(t *testing.T) {
	for _, c := range []struct{ el, body, want string }{
		{"msup", `<mml:msup><mml:mi>x</mml:mi></mml:msup>`, "x"},
		{"msub", `<mml:msub><mml:mi>x</mml:mi><mml:mn>1</mml:mn><mml:mn>2</mml:mn></mml:msub>`, "x12"},
		{"mfrac", `<mml:mfrac><mml:mi>x</mml:mi></mml:mfrac>`, "x"},
		{"mroot", `<mml:mroot><mml:mi>x</mml:mi></mml:mroot>`, "x"},
		{"mfrac", `<mml:mfrac><mml:mi>x</mml:mi><mml:mi>y</mml:mi><mml:mi>z</mml:mi></mml:mfrac>`, "xyz"},
		{"mroot", `<mml:mroot><mml:mi>x</mml:mi><mml:mn>3</mml:mn><mml:mi>z</mml:mi></mml:mroot>`, "x3z"},
	} {
		got, ws := PlainText("a " + mml(c.body) + " b")
		if got != "a "+c.want+" b" || len(ws) != 1 || ws[0].Code != WarnMathUnconverted || ws[0].Fragment != c.el {
			t.Errorf("%s: got %q, warnings %+v", c.el, got, ws)
		}
	}
}

// TestMathML_TokenTable checks a whole-token table entry wins, and that the
// table also applies to text kept from an unconverted element.
func TestMathML_TokenTable(t *testing.T) {
	sym, err := DefaultSymbols()
	if err != nil {
		t.Fatal(err)
	}
	sym.Tokens["ab"] = "Z"
	got, _ := plainText(mml(`<mml:mi>ab</mml:mi>`), sym)
	if got != "Z" {
		t.Errorf("whole-token entry: got %q, want Z", got)
	}
	got, ws := PlainText(mml("<mml:mover><mml:mi>x\u2062</mml:mi><mml:mo>¯</mml:mo></mml:mover>"))
	if got != "x¯" || len(ws) != 1 {
		t.Errorf("table in an unconverted element: got %q, warnings %+v", got, ws)
	}
}

// TestMathML_LinearNesting checks a linear form counts as more than one token
// when it is an operand, and that a long unmappable root index is bracketed.
func TestMathML_LinearNesting(t *testing.T) {
	got, ws := PlainText("a " + mml(`<mml:mfrac><mml:msup><mml:mi>x</mml:mi><mml:mi>q</mml:mi></mml:msup><mml:mn>2</mml:mn></mml:mfrac>`) + " b")
	if got != "a (x^q)/2 b" || len(ws) != 1 || ws[0].Code != WarnMathLinear {
		t.Errorf("linear operand: got %q, warnings %+v", got, ws)
	}
	got, ws = PlainText("a " + mml(`<mml:mroot><mml:mi>x</mml:mi><mml:mrow><mml:mi>q</mml:mi><mml:mo>+</mml:mo><mml:mn>1</mml:mn></mml:mrow></mml:mroot>`) + " b")
	if got != "a x^(1/(q+1)) b" || len(ws) != 1 || ws[0].Code != WarnMathLinear {
		t.Errorf("long linear index: got %q, warnings %+v", got, ws)
	}
}

// TestMathML_TokenSplitByComment checks a token whose text a comment splits in
// two is read whole.
func TestMathML_TokenSplitByComment(t *testing.T) {
	got, ws := PlainText("a " + mml(`<mml:mi>x<!-- note -->y</mml:mi>`) + " b")
	if got != "a xy b" || len(ws) != 0 {
		t.Errorf("got %q, warnings %+v", got, ws)
	}
}
