package crossrefapi

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// mathPiece is the Unicode text of part of a formula and how many MathML
// tokens it holds; a piece of more than one token is bracketed when it is the
// operand of a fraction, root or square root.
type mathPiece struct {
	text   string
	tokens int
}

// mathUnicode converts a MathML formula to Unicode text for a title. Tokens
// (mi, mn, mo, mtext) become their text, with the symbol table's token
// replacements applied; mrow, mstyle and mpadded join their children; msup and
// msub become Unicode super and subscripts when every character has one, and a
// linear form ("x^q", "Z_N") otherwise; mfrac becomes "a/b", msqrt "√x" and
// mroot "ⁿ√x", with brackets around an operand of more than one token. Any
// other element is reduced to its text content. Warnings say which.
func mathUnicode(n *node, sym *Symbols, warnings *[]Warning) string {
	return convertMath(n, sym, warnings).text
}

// convertMath converts one node of a formula.
func convertMath(n *node, sym *Symbols, warnings *[]Warning) mathPiece {
	switch n.name {
	case "":
		return tokenPiece(n.text, sym)
	case "math", "mrow", "mstyle", "mpadded":
		return joinPieces(convertKids(n, sym, warnings))
	case "mi", "mn", "mo", "mtext", "ms":
		return tokenPiece(allText(n), sym)
	case "msup", "msub":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return unconverted(n, sym, warnings)
		}
		base := convertMath(kids[0], sym, warnings)
		script := convertMath(kids[1], sym, warnings)
		table, op := sym.Subscripts, "_"
		if n.name == "msup" {
			table, op = sym.Superscripts, "^"
		}
		if mapped, ok := mapAll(script.text, table); ok {
			return mathPiece{base.text + mapped, base.tokens + script.tokens}
		}
		return linear(base.text+op+bracketRunes(script.text), base.tokens+script.tokens, warnings)
	case "mfrac":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return unconverted(n, sym, warnings)
		}
		num := convertMath(kids[0], sym, warnings)
		den := convertMath(kids[1], sym, warnings)
		return mathPiece{bracket(num) + "/" + bracket(den), num.tokens + den.tokens + 1}
	case "msqrt":
		content := joinPieces(convertKids(n, sym, warnings))
		return mathPiece{"√" + bracket(content), content.tokens + 1}
	case "mroot":
		kids := mathChildren(n)
		if len(kids) != 2 {
			return unconverted(n, sym, warnings)
		}
		base := convertMath(kids[0], sym, warnings)
		index := convertMath(kids[1], sym, warnings)
		if mapped, ok := mapAll(index.text, sym.Superscripts); ok {
			return mathPiece{mapped + "√" + bracket(base), base.tokens + index.tokens + 1}
		}
		return linear(bracket(base)+"^(1/"+bracket(index)+")", base.tokens+index.tokens+1, warnings)
	}
	return unconverted(n, sym, warnings)
}

// convertKids converts the children of a node.
func convertKids(n *node, sym *Symbols, warnings *[]Warning) []mathPiece {
	var out []mathPiece
	for _, k := range mathChildren(n) {
		out = append(out, convertMath(k, sym, warnings))
	}
	return out
}

// mathChildren returns the children of a node that carry content: elements,
// and text that is not just layout whitespace.
func mathChildren(n *node) []*node {
	var out []*node
	for _, k := range n.kids {
		if k.name == "" && strings.TrimSpace(k.text) == "" {
			continue
		}
		out = append(out, k)
	}
	return out
}

// joinPieces concatenates pieces.
func joinPieces(ps []mathPiece) mathPiece {
	var b strings.Builder
	tokens := 0
	for _, p := range ps {
		b.WriteString(p.text)
		tokens += p.tokens
	}
	return mathPiece{b.String(), tokens}
}

// bracket wraps a piece of more than one token in parentheses.
func bracket(p mathPiece) string {
	if p.tokens > 1 {
		return "(" + p.text + ")"
	}
	return p.text
}

// bracketRunes wraps a script of more than one character in parentheses, for
// the linear form: "x^q" but "x^(2q)".
func bracketRunes(s string) string {
	if utf8.RuneCountInString(s) > 1 {
		return "(" + s + ")"
	}
	return s
}

// mapAll maps every character of s through table. It reports false when any
// character has no entry; an empty s maps to an empty string.
func mapAll(s string, table map[string]string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		v, ok := table[string(r)]
		if !ok {
			return "", false
		}
		b.WriteString(v)
	}
	return b.String(), true
}

// linear returns a formula written in its linear form and records the warning.
func linear(text string, tokens int, warnings *[]Warning) mathPiece {
	*warnings = append(*warnings, Warning{Code: WarnMathLinear, Fragment: text})
	return mathPiece{text, tokens}
}

// unconverted reduces an element with no conversion to its text content and
// records which element it was.
func unconverted(n *node, sym *Symbols, warnings *[]Warning) mathPiece {
	*warnings = append(*warnings, Warning{Code: WarnMathUnconverted, Fragment: n.name})
	var ps []mathPiece
	var walk func(*node)
	walk = func(n *node) {
		if n.name == "" {
			if p := tokenPiece(n.text, sym); p.tokens > 0 {
				ps = append(ps, p)
			}
			return
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(n)
	return joinPieces(ps)
}

// allText returns all the text under a node.
func allText(n *node) string {
	if n.name == "" {
		return n.text
	}
	var b strings.Builder
	for _, k := range n.kids {
		b.WriteString(allText(k))
	}
	return b.String()
}

// tokenPiece turns the text of a token into a piece: whitespace collapsed, then
// the symbol table's token replacements applied, the whole token first and
// then character by character. A token that ends up empty holds no tokens.
func tokenPiece(text string, sym *Symbols) mathPiece {
	text = strings.Join(strings.FieldsFunc(text, unicode.IsSpace), " ")
	if v, ok := sym.Tokens[text]; ok {
		text = v
	} else {
		var b strings.Builder
		for _, r := range text {
			if v, ok := sym.Tokens[string(r)]; ok {
				b.WriteString(v)
			} else {
				b.WriteRune(r)
			}
		}
		text = b.String()
	}
	if text == "" {
		return mathPiece{}
	}
	return mathPiece{text, 1}
}
