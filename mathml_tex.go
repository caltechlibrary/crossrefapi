package crossrefapi

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// texPiece is the TeX of part of a formula and how many MathML tokens it holds;
// a piece of more than one token is put in braces when it is the base of a
// script.
type texPiece struct {
	text   string
	tokens int
}

// letterCommandEnd matches a TeX command named with letters at the end of a
// string: such a command needs a space before a letter or digit that follows.
var letterCommandEnd = regexp.MustCompile(`\\[A-Za-z]+$`)

// mathTeX converts a MathML formula to TeX (without the surrounding math
// delimiters). Tokens become TeX through the symbol table (Greek letters and
// operators to commands; a multi-letter identifier to \mathrm{}; a mathvariant
// to its font command; mtext to \text{}), scripts become ^{} and _{}, fractions
// and roots \frac{}{}, \sqrt{}, \sqrt[]{}, and an over-accent a command from
// the table's accents. Any other element is reduced to its text inside \text{}
// with a warning.
func mathTeX(n *node, sym *Symbols, warnings *[]Warning) string {
	return texNode(n, sym, warnings).text
}

// texNode converts one node of a formula.
func texNode(n *node, sym *Symbols, warnings *[]Warning) texPiece {
	switch n.name {
	case "":
		return texToken(n.text, sym, texEscape)
	case "math", "mrow", "mstyle", "mpadded":
		return joinTeX(texKids(n, sym, warnings))
	case "mi":
		return texIdentifier(n, sym)
	case "mn":
		return texToken(allText(n), sym, texEscape)
	case "mo":
		return texToken(allText(n), sym, func(t string) string {
			if cmd, ok := sym.TeX[t]; ok {
				return cmd
			}
			return texEscape(t)
		})
	case "mtext", "ms":
		return texToken(allText(n), sym, func(t string) string { return texText(t, sym) })
	case "mspace":
		return texPiece{"~", 1}
	case "msup", "msub":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return texUnconverted(n, sym, warnings)
		}
		op := "_"
		if n.name == "msup" {
			op = "^"
		}
		base := texNode(kids[0], sym, warnings)
		script := texNode(kids[1], sym, warnings)
		return texPiece{texBase(base) + op + "{" + script.text + "}", base.tokens + script.tokens}
	case "msubsup", "munderover":
		kids := mathChildren(n)
		if len(kids) != 3 {
			return texUnconverted(n, sym, warnings)
		}
		base := texNode(kids[0], sym, warnings)
		low := texNode(kids[1], sym, warnings)
		high := texNode(kids[2], sym, warnings)
		return texPiece{texBase(base) + "_{" + low.text + "}^{" + high.text + "}", base.tokens + low.tokens + high.tokens}
	case "munder":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return texUnconverted(n, sym, warnings)
		}
		base := texNode(kids[0], sym, warnings)
		under := texNode(kids[1], sym, warnings)
		return texPiece{texBase(base) + "_{" + under.text + "}", base.tokens + under.tokens}
	case "mover":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return texUnconverted(n, sym, warnings)
		}
		base := texNode(kids[0], sym, warnings)
		if over := onlyChild(kids[1]); over.name == "mo" {
			if cmd, ok := sym.Accents[tokenPiece(allText(over), sym).text]; ok {
				return texPiece{cmd + "{" + base.text + "}", 1}
			}
		}
		over := texNode(kids[1], sym, warnings)
		return texPiece{texBase(base) + "^{" + over.text + "}", base.tokens + over.tokens}
	case "mfrac":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return texUnconverted(n, sym, warnings)
		}
		num := texNode(kids[0], sym, warnings)
		den := texNode(kids[1], sym, warnings)
		return texPiece{`\frac{` + num.text + `}{` + den.text + `}`, num.tokens + den.tokens}
	case "msqrt":
		content := joinTeX(texKids(n, sym, warnings))
		return texPiece{`\sqrt{` + content.text + `}`, content.tokens}
	case "mroot":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return texUnconverted(n, sym, warnings)
		}
		base := texNode(kids[0], sym, warnings)
		index := texNode(kids[1], sym, warnings)
		return texPiece{`\sqrt[` + index.text + `]{` + base.text + `}`, base.tokens + index.tokens}
	}
	return texUnconverted(n, sym, warnings)
}

// texKids converts the children of a node.
func texKids(n *node, sym *Symbols, warnings *[]Warning) []texPiece {
	var out []texPiece
	for _, k := range mathChildren(n) {
		out = append(out, texNode(k, sym, warnings))
	}
	return out
}

// onlyChild unwraps a row of one child, so an accent inside an mrow is found.
func onlyChild(n *node) *node {
	for n.name == "mrow" {
		kids := mathChildren(n)
		if len(kids) != 1 {
			break
		}
		n = kids[0]
	}
	return n
}

// joinTeX concatenates pieces, putting a space after a letter command when a
// letter or digit follows.
func joinTeX(ps []texPiece) texPiece {
	var b strings.Builder
	tokens := 0
	for _, p := range ps {
		if p.text == "" {
			continue
		}
		cur := b.String()
		if first, _ := utf8.DecodeRuneInString(p.text); letterCommandEnd.MatchString(cur) && (unicode.IsLetter(first) || unicode.IsDigit(first)) {
			b.WriteByte(' ')
		}
		b.WriteString(p.text)
		tokens += p.tokens
	}
	return texPiece{b.String(), tokens}
}

// texBase returns the base of a script: empty becomes {} so the script has
// something to attach to, and a base of more than one token is put in braces.
func texBase(p texPiece) string {
	switch {
	case p.text == "":
		return "{}"
	case p.tokens > 1:
		return "{" + p.text + "}"
	}
	return p.text
}

// texToken applies the table's token replacements to a token's text and, if
// anything is left, converts it with conv.
func texToken(text string, sym *Symbols, conv func(string) string) texPiece {
	t := tokenPiece(text, sym)
	if t.tokens == 0 {
		return texPiece{}
	}
	return texPiece{conv(t.text), 1}
}

// texIdentifier converts an mi: one character through the table (a Greek letter
// to its command), several characters to \mathrm{}, and a mathvariant from the
// table to its font command around the result.
func texIdentifier(n *node, sym *Symbols) texPiece {
	t := tokenPiece(allText(n), sym)
	if t.tokens == 0 {
		return texPiece{}
	}
	variant := sym.MathVariants[n.attrs["mathvariant"]]
	if utf8.RuneCountInString(t.text) == 1 {
		body := texEscape(t.text)
		if cmd, ok := sym.TeX[t.text]; ok {
			body = cmd
		}
		if variant != "" {
			body = variant + "{" + body + "}"
		}
		return texPiece{body, 1}
	}
	if variant == "" {
		variant = `\mathrm`
	}
	return texPiece{variant + "{" + texEscape(t.text) + "}", 1}
}

// texText converts the text of an mtext: one character with a table entry
// becomes that command, anything else \text{}.
func texText(t string, sym *Symbols) string {
	if utf8.RuneCountInString(t) == 1 {
		if cmd, ok := sym.TeX[t]; ok {
			return cmd
		}
	}
	return `\text{` + texEscape(t) + `}`
}

// texEscape escapes the characters TeX treats specially.
func texEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\textbackslash{}`)
		case '{', '}', '$', '&', '#', '%', '_':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// texUnconverted reduces an element with no conversion to its text inside
// \text{} and records which element it was.
func texUnconverted(n *node, sym *Symbols, warnings *[]Warning) texPiece {
	*warnings = append(*warnings, Warning{Code: WarnMathUnconverted, Fragment: n.name})
	text := contentText(n, sym)
	if text == "" {
		return texPiece{}
	}
	return texPiece{`\text{` + texEscape(text) + `}`, 1}
}
