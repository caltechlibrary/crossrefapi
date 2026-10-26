package crossrefapi

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// mathMLNamespace is the XML namespace of MathML. A formula is recognized by
// it, never by a prefix: CrossRef uses "mml" in some records and a different
// one-letter prefix per formula in others.
const mathMLNamespace = "http://www.w3.org/1998/Math/MathML"

// Warning codes returned in [Warning.Code].
const (
	// WarnParseFallback means the fragment is not well-formed markup. The
	// text was recovered by stripping tags and decoding entities, and inline
	// whitespace rules were not applied.
	WarnParseFallback = "parse-fallback"
	// WarnScriptUnmapped means a sub or superscript has a character with no
	// Unicode form in the symbol table; its text was kept as it was.
	WarnScriptUnmapped = "script-unmapped"
	// WarnMathLinear means a formula had a script or root the Unicode forms
	// cannot show, so it was written in a linear form such as "Z_N".
	WarnMathLinear = "math-linear"
	// WarnMathUnconverted means a MathML element has no conversion yet; the
	// formula was reduced to its text content. Fragment is the element name.
	WarnMathUnconverted = "math-unconverted"
)

// Warning describes something the text normalizer could not convert cleanly.
// The text it returns is still the best available; a warning lets the caller
// show the record to a person.
//
// Code is a short stable identifier (see the Warn constants), and Fragment is
// the markup or text the warning is about.
type Warning struct {
	Code     string
	Fragment string
}

// String returns the warning as "code: fragment".
func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.Code, w.Fragment)
}

// defaultSymbolsOnce loads the embedded table once. The table is only read
// afterwards, so it is safe to share between goroutines.
var defaultSymbolsOnce = sync.OnceValues(DefaultSymbols)

var (
	entityRefRE   = regexp.MustCompile(`&([A-Za-z][A-Za-z0-9]*);`)
	anyEntityRE   = regexp.MustCompile(`&(?:#[0-9]+|#[xX][0-9a-fA-F]+|[A-Za-z][A-Za-z0-9]*);`)
	predefinedXML = map[string]bool{"lt": true, "gt": true, "amp": true, "quot": true, "apos": true}
)

// PlainText converts a CrossRef title (or other short inline fragment) to
// plain Unicode text. It removes tags and keeps their text, decodes character
// entities, maps <sub> and <sup> to Unicode super and subscripts when every
// character has one, and tidies whitespace.
//
// Whitespace follows decision A of crossrefapi/DR-0001. Publishers pretty-print
// markup, so a newline and indentation can appear at every tag boundary and the
// original gap is lost. Whitespace that contains a newline and touches a <sub>,
// a <sup> or a one-character italic is dropped; whitespace beside any other
// tag collapses to one space. A space is inserted where a formula abuts a word,
// and a space between a formula and a following hyphen is removed. Non-breaking
// and thin spaces count as whitespace.
//
// A formula (MathML, recognized by its namespace) is converted to Unicode text
// (see mathUnicode): scripts become Unicode super and subscripts, fractions
// "a/b", roots "√x". Anything that cannot be shown that way is written in a
// linear form or reduced to its text, with a warning. A fragment that is not well-formed (such as a raw "<") is
// still converted, with a WarnParseFallback warning. The result is plain text:
// do not feed it back in if it contains "<".
//
// @param s {string} a CrossRef title, subtitle or other inline fragment
// @returns {string} the plain text, with no leading or trailing whitespace
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	title, warnings := crossrefapi.PlainText("<i>Euclid</i>\n   preparation")
//	fmt.Println(title, len(warnings)) // Euclid preparation 0
func PlainText(s string) (string, []Warning) {
	sym, err := defaultSymbolsOnce()
	if err != nil {
		// The embedded table is checked by this package's tests; reaching
		// this means the package was built with a damaged symbols.json.
		panic("crossrefapi: embedded symbols.json is malformed: " + err.Error())
	}
	return plainText(s, sym)
}

// plainText is PlainText with an explicit symbol table.
func plainText(s string, sym *Symbols) (string, []Warning) {
	root, err := parseFragment(s, sym)
	if err != nil {
		return fallbackText(s, sym), []Warning{{Code: WarnParseFallback, Fragment: s}}
	}
	var warnings []Warning
	return resolveEvents(flatten(root.kids, sym, &warnings)), warnings
}

// node is one element or text run of a parsed fragment.
type node struct {
	space string // namespace (or an undeclared prefix) of an element
	name  string // lower-cased local name; "" for a text node
	text  string // the text of a text node
	kids  []*node
}

// parseFragment parses s as inline markup under a synthetic root. The parser
// is deliberately forgiving: it invents missing end tags, leaves unknown
// entities and bare ampersands alone, and treats an undeclared prefix such as
// "jats:" as part of the name.
func parseFragment(s string, sym *Symbols) (*node, error) {
	d := xml.NewDecoder(strings.NewReader("<root>" + s + "</root>"))
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = entityMap(s, sym)
	root := &node{name: "root"}
	stack := []*node{root}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		top := stack[len(stack)-1]
		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{space: t.Name.Space, name: strings.ToLower(t.Name.Local)}
			top.kids = append(top.kids, n)
			stack = append(stack, n)
		case xml.EndElement:
			// The decoder reports an end tag with no start tag itself, so this
			// cannot happen today; it stays so a change in the decoder cannot make
			// the pop below fail.
			if len(stack) == 1 {
				return nil, fmt.Errorf("unbalanced end element %q", t.Name.Local)
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			top.kids = append(top.kids, &node{text: string(t)})
		}
	}
	return root, nil
}

// entityMap resolves the named character entities used in s that XML does not
// predefine: from the symbol table first, then from the HTML entity tables.
// Resolving them here, as the parser's entity map, decodes each reference once.
func entityMap(s string, sym *Symbols) map[string]string {
	m := map[string]string{}
	for _, ref := range entityRefRE.FindAllStringSubmatch(s, -1) {
		name := ref[1]
		if predefinedXML[name] {
			continue
		}
		if v, ok := sym.Entities[name]; ok {
			m[name] = v
		} else if v := html.UnescapeString(ref[0]); v != ref[0] {
			m[name] = v
		}
	}
	return m
}

// fallbackText recovers text from a fragment the parser rejected: real tags
// (a "<" followed by a letter, "/", "!" or "?" and a closing ">") are dropped,
// every other "<" is kept as text, entities are decoded once, and whitespace is
// collapsed.
func fallbackText(s string, sym *Symbols) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '<' && i+1 < len(s) && (isASCIILetter(s[i+1]) || strings.IndexByte("/!?", s[i+1]) >= 0) {
			if j := strings.IndexByte(s[i:], '>'); j >= 0 {
				i += j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	text := anyEntityRE.ReplaceAllStringFunc(b.String(), func(ref string) string {
		if name := ref[1 : len(ref)-1]; !strings.HasPrefix(name, "#") {
			if v, ok := sym.Entities[name]; ok {
				return v
			}
		}
		return html.UnescapeString(ref)
	})
	return strings.Join(strings.FieldsFunc(text, unicode.IsSpace), " ")
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// event kinds for the flattened fragment.
const (
	evText     = iota // a run of text, whitespace included
	evBoundary        // a tag boundary that matters to whitespace (glue)
	evAtom            // text that is never split or trimmed: a script or a formula
)

// event is one step of the flattened fragment that resolveEvents turns into text.
type event struct {
	kind int
	text string
	glue bool // behaves like a sub or superscript beside layout whitespace
	math bool // the atom is a formula
}

// flatten turns parsed nodes into events. Scripts and formulas become atoms;
// a one-character italic becomes a pair of glue boundaries around its text;
// every other element just contributes its children.
func flatten(nodes []*node, sym *Symbols, warnings *[]Warning) []event {
	var ev []event
	for _, n := range nodes {
		switch {
		case n.name == "":
			ev = append(ev, event{kind: evText, text: n.text})
		case n.space == mathMLNamespace || n.name == "math":
			if t := mathUnicode(n, sym, warnings); t != "" {
				ev = append(ev, event{kind: evAtom, text: t, math: true})
			}
		case n.name == "br":
			ev = append(ev, event{kind: evText, text: "\n"})
		case n.name == "sub" || n.name == "sup":
			if a, ok := scriptAtom(n, sym, warnings); ok {
				ev = append(ev, a)
			}
		case n.name == "i" || n.name == "em" || n.name == "italic":
			inner := innerText(n, sym, nil)
			glue := utf8.RuneCountInString(inner) == 1
			if glue {
				ev = append(ev, event{kind: evBoundary, glue: true})
			}
			ev = append(ev, flatten(n.kids, sym, warnings)...)
			if glue {
				ev = append(ev, event{kind: evBoundary, glue: true})
			}
		default:
			ev = append(ev, flatten(n.kids, sym, warnings)...)
		}
	}
	return ev
}

// innerText is the resolved text of an element's children.
func innerText(n *node, sym *Symbols, warnings *[]Warning) string {
	var discard []Warning
	if warnings == nil {
		warnings = &discard
	}
	return resolveEvents(flatten(n.kids, sym, warnings))
}

// scriptAtom builds the atom for a <sub> or <sup>: its text in Unicode
// super or subscript characters when every character has one, otherwise the
// text as it was, with a warning.
func scriptAtom(n *node, sym *Symbols, warnings *[]Warning) (event, bool) {
	var nested []Warning
	inner := resolveEvents(flatten(n.kids, sym, &nested))
	*warnings = append(*warnings, nested...)
	if inner == "" {
		return event{}, false
	}
	table := sym.Subscripts
	if n.name == "sup" {
		table = sym.Superscripts
	}
	var mapped strings.Builder
	for _, r := range inner {
		v, ok := table[string(r)]
		if !ok {
			*warnings = append(*warnings, Warning{
				Code:     WarnScriptUnmapped,
				Fragment: "<" + n.name + ">" + inner + "</" + n.name + ">",
			})
			return event{kind: evAtom, text: inner, glue: true}, true
		}
		mapped.WriteString(v)
	}
	return event{kind: evAtom, text: mapped.String(), glue: true}, true
}

// resolveEvents joins events into text, deciding every whitespace gap.
//
// A gap is the whitespace between two pieces of text, however many tags lie in
// it. It is dropped at the start, and when it contains a newline and touches
// glue (a script or a one-character italic); dropped before a hyphen that
// follows a formula; and otherwise written as one space. With no gap at all, a
// space is inserted between a formula and a word it abuts.
func resolveEvents(ev []event) string {
	var out strings.Builder
	var last rune
	type gapState struct{ ws, nl, glue bool }
	var gap gapState
	prevGlue := false // the last event since the last text was glue
	prevMath := false // the last piece written was a formula

	flush := func(piece string, pieceMath bool) {
		first, _ := utf8.DecodeRuneInString(piece)
		switch {
		case gap.ws:
			drop := out.Len() == 0 ||
				(gap.nl && gap.glue) ||
				(prevMath && isHyphen(first))
			if !drop {
				out.WriteByte(' ')
			}
		case out.Len() > 0 && (prevMath || pieceMath) && isAlnum(last) && isAlnum(first):
			out.WriteByte(' ')
		}
		gap = gapState{}
	}
	write := func(piece string, math, glue bool) {
		flush(piece, math)
		out.WriteString(piece)
		last, _ = utf8.DecodeLastRuneInString(piece)
		prevMath = math
		prevGlue = glue
	}

	for _, e := range ev {
		switch e.kind {
		case evBoundary:
			if gap.ws && e.glue {
				gap.glue = true
			}
			if e.glue {
				prevGlue = true
			}
		case evAtom:
			if gap.ws && e.glue {
				gap.glue = true
			}
			write(e.text, e.math, e.glue)
		case evText:
			rest := e.text
			for rest != "" {
				isSpace := unicode.IsSpace(firstRune(rest))
				n := strings.IndexFunc(rest, func(r rune) bool { return unicode.IsSpace(r) != isSpace })
				if n < 0 {
					n = len(rest)
				}
				run := rest[:n]
				rest = rest[n:]
				if isSpace {
					gap.ws = true
					gap.nl = gap.nl || strings.ContainsAny(run, "\n\r")
					gap.glue = gap.glue || prevGlue
				} else {
					write(run, false, false)
				}
			}
		}
	}
	return out.String()
}

func firstRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

func isHyphen(r rune) bool { return r == '-' || r == '\u2010' }

// isAlnum reports a letter or any kind of number, so a superscript digit
// (²) counts as part of a word.
func isAlnum(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }
