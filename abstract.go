package crossrefapi

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// HTMLAbstract converts a CrossRef abstract (JATS markup, sometimes plain HTML
// or plain text) to HTML that RDM's sanitizer keeps: paragraphs, headings,
// lists, tables, links and the inline tags it allows, with formulas as TeX in
// \( ... \) for MathJax. Everything else is dropped, and what is dropped is
// documented here so a record can be checked against it.
//
// Structure: <jats:p> becomes <p>; <jats:title> becomes <p><b>...</b></p>,
// except the abstract's first heading when it is in the symbol table's list of
// dropped headings (by default "Abstract", ignoring case and a trailing colon;
// crossrefapi/DR-0001 decision B). <jats:sec> and other unknown elements are
// unwrapped, keeping their text; <jats:supplementary-material> and
// <jats:permissions> (copyright statements) are dropped with their content, as
// are <script> and <style>.
//
// Inline: <jats:italic> and <jats:bold> become <i> and <b>; <sub>, <sup> and the
// other allowed tags are kept; a link keeps only an http, https, mailto or
// relative href and its title; every other attribute is dropped.
//
// Formulas: MathML, recognized by its namespace, becomes TeX. The publisher's
// own <jats:tex-math> is ignored, except that a formula with no MathML at all
// uses it. Whitespace follows crossrefapi/DR-0001 decision A, so layout
// newlines beside a script or a one-letter italic are dropped (which can turn
// "(4 < M < 8)" into "(4 <M< 8)"), and none is left between blocks.
//
// A fragment that is not well-formed (a raw "<") is still converted, as plain
// escaped text, with a parse-fallback warning. Warnings from formulas that could
// not be converted are returned too. The output is HTML: do not pass it to
// PlainText.
//
// @param s {string} a CrossRef abstract
// @returns {string} sanitizer-safe HTML; empty when the abstract has no text
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	html, warnings := crossrefapi.HTMLAbstract("<jats:p>Water is H<jats:sub>2</jats:sub>O.</jats:p>")
//	fmt.Println(html, len(warnings)) // <p>Water is H<sub>2</sub>O.</p> 0
func HTMLAbstract(s string) (string, []Warning) {
	return htmlAbstract(s, defaultTable())
}

// htmlAbstract is HTMLAbstract with an explicit symbol table.
func htmlAbstract(s string, sym *Symbols) (string, []Warning) {
	root, err := parseFragment(s, sym)
	if err != nil {
		return htmlEscape(fallbackText(s, sym)), []Warning{{Code: WarnParseFallback, Fragment: s}}
	}
	var warnings []Warning
	st := &abstractState{sym: sym, warnings: &warnings, dropped: map[string]bool{}}
	for _, h := range sym.DroppedHeadings {
		st.dropped[normalizeHeading(h)] = true
	}
	return strings.TrimSpace(resolveEvents(st.flatten(root.kids), htmlEscape)), warnings
}

// abstractState carries what the conversion has seen so far.
type abstractState struct {
	sym        *Symbols
	warnings   *[]Warning
	dropped    map[string]bool // normalized headings to drop when first
	sawContent bool            // text or a formula came before this point
	sawHeading bool            // a non-empty heading came before this point
}

// blockTags maps a source element to its HTML block tag and whether an empty
// one is skipped. Table and list structure keeps its empty cells and items.
var blockTags = map[string]struct {
	tag       string
	skipEmpty bool
}{
	"p": {"p", true}, "h1": {"h1", true}, "h2": {"h2", true}, "h3": {"h3", true},
	"h4": {"h4", true}, "h5": {"h5", true}, "blockquote": {"blockquote", true},
	"pre": {"pre", true}, "div": {"div", true},
	"ul": {"ul", false}, "ol": {"ol", false}, "li": {"li", false},
	"table": {"table", false}, "thead": {"thead", false}, "tbody": {"tbody", false},
	"tfoot": {"tfoot", false}, "tr": {"tr", false}, "th": {"th", false}, "td": {"td", false},
}

// inlineTags maps a source element to its HTML inline tag.
var inlineTags = map[string]string{
	"i": "i", "italic": "i", "em": "em", "b": "b", "bold": "b", "strong": "strong",
	"u": "u", "s": "s", "strike": "strike", "code": "code", "span": "span",
	"abbr": "abbr", "acronym": "acronym", "a": "a", "sub": "sub", "sup": "sup",
}

// flatten turns parsed nodes into events for the HTML output.
func (st *abstractState) flatten(nodes []*node) []event {
	var ev []event
	for _, n := range nodes {
		switch {
		case n.name == "":
			if strings.TrimSpace(n.text) != "" {
				st.sawContent = true
			}
			ev = append(ev, event{kind: evText, text: n.text})
		case n.space == mathMLNamespace || n.name == "math":
			ev = append(ev, st.formula(n)...)
		default:
			ev = append(ev, st.element(n)...)
		}
	}
	return ev
}

// formula converts a MathML formula to a TeX atom. Its spacing is judged by
// the Unicode form of the formula, which is what a reader sees at its edges.
func (st *abstractState) formula(n *node) []event {
	tex := mathTeX(n, st.sym, st.warnings)
	if tex == "" {
		return nil
	}
	var scratch []Warning
	st.sawContent = true
	return []event{{kind: evAtom, text: htmlEscape(`\(` + tex + `\)`), vis: mathUnicode(n, st.sym, &scratch), math: true}}
}

// element converts one element.
func (st *abstractState) element(n *node) []event {
	switch n.name {
	case "supplementary-material", "permissions", "tex-math", "script", "style":
		return nil
	case "inline-formula", "alternatives", "disp-formula":
		if !hasMath(n) {
			tex := strings.Join(strings.Fields(stripTeXDelims(firstTeXMath(n))), " ")
			switch {
			case tex == "":
			case !safeTeX(tex, st.sym):
				// Publisher TeX is untrusted input that a page's math renderer
				// will run: a link, a macro or a delimiter would do harm.
				*st.warnings = append(*st.warnings, Warning{Code: WarnMathUnconverted, Fragment: "tex-math"})
			default:
				st.sawContent = true
				return []event{{kind: evAtom, text: htmlEscape(`\(` + tex + `\)`), vis: tex, math: true}}
			}
			return nil
		}
		return st.flatten(n.kids)
	case "title":
		return st.heading(n)
	case "br":
		return []event{{kind: evTag, raw: "<br>", block: true}}
	}
	if b, ok := blockTags[n.name]; ok {
		inner := st.flatten(n.kids)
		if b.skipEmpty && !hasContent(inner) {
			return nil
		}
		return wrap(b.tag+blockAttrs(n), "</"+b.tag+">", inner, true, false)
	}
	if tag, ok := inlineTags[n.name]; ok {
		attrs, keep := inlineAttrs(tag, n)
		if !keep {
			return st.flatten(n.kids)
		}
		inner := st.flatten(n.kids)
		if !hasContent(inner) {
			return nil
		}
		glue := tag == "sub" || tag == "sup" ||
			((tag == "i" || tag == "em") && utf8.RuneCountInString(strings.TrimSpace(allText(n))) == 1)
		return wrap(tag+attrs, "</"+tag+">", inner, false, glue)
	}
	return st.flatten(n.kids)
}

// heading converts a <jats:title>: a bold paragraph, unless it is the
// abstract's first heading and the symbol table lists it as one to drop. An
// empty heading is skipped and does not count as the first.
func (st *abstractState) heading(n *node) []event {
	text := strings.Join(strings.Fields(allText(n)), " ")
	if text == "" {
		return nil
	}
	leading := !st.sawContent && !st.sawHeading
	st.sawHeading = true
	if leading && st.dropped[normalizeHeading(text)] {
		return nil
	}
	inner := st.flatten(n.kids)
	ev := []event{{kind: evTag, raw: "<p>", block: true}, {kind: evTag, raw: "<b>"}}
	ev = append(ev, inner...)
	return append(ev, event{kind: evTag, raw: "</b>"}, event{kind: evTag, raw: "</p>", block: true})
}

// wrap puts an open and a close tag around events.
func wrap(open, close string, inner []event, block, glue bool) []event {
	ev := make([]event, 0, len(inner)+2)
	ev = append(ev, event{kind: evTag, raw: "<" + open + ">", block: block, glue: glue})
	ev = append(ev, inner...)
	return append(ev, event{kind: evTag, raw: close, block: block, glue: glue})
}

// hasContent reports whether events hold any text or formula.
func hasContent(ev []event) bool {
	for _, e := range ev {
		if e.kind == evAtom || (e.kind == evText && strings.TrimSpace(e.text) != "") {
			return true
		}
	}
	return false
}

// hasMath reports whether a MathML formula is under a node.
func hasMath(n *node) bool {
	for _, k := range n.kids {
		if k.name != "" && (k.space == mathMLNamespace || k.name == "math" || hasMath(k)) {
			return true
		}
	}
	return false
}

// firstTeXMath returns the text of the first <tex-math> under a node.
func firstTeXMath(n *node) string {
	for _, k := range n.kids {
		if k.name == "tex-math" {
			return allText(k)
		}
		if k.name != "" {
			if t := firstTeXMath(k); t != "" {
				return t
			}
		}
	}
	return ""
}

// texControlRE matches a TeX control sequence: a backslash and a name of
// letters, or a backslash and one other character.
var texControlRE = regexp.MustCompile(`\\([A-Za-z]+|[^A-Za-z])`)

// texCommandNameRE finds the command names in a TeX command string.
var texCommandNameRE = regexp.MustCompile(`\\([A-Za-z]+)`)

// safeTeXEscapes are the control symbols safe TeX may use: escaped specials and
// spacing. The delimiters \( \) \[ \] are not among them, since they would end
// the formula early.
const safeTeXEscapes = "%&#_{}$,;:! |\\"

// safeTeXCommands are the TeX commands publisher TeX may use. They are all
// mathematics: letters, operators, fonts, accents, fractions, roots, limits and
// spacing. Links (\href, \url), macros and definitions (\def, \newcommand,
// \let), loading (\require, \input), styling (\class, \style, \cssId),
// raw characters (\unicode, \char) and environments (\begin, \end) are not.
// The commands the symbol table itself produces are allowed as well.
var safeTeXCommands = func() map[string]bool {
	m := map[string]bool{}
	for _, c := range strings.Fields(`frac dfrac tfrac binom choose atop sqrt text textbf textit textrm
		mathrm mathit mathbf mathbb mathcal mathfrak mathsf mathtt mathop mathbin mathrel mathord
		boldsymbol operatorname left right big Big bigg Bigg bigl bigr Bigl Bigr overline underline
		widehat widetilde overbrace underbrace hat tilde vec bar dot ddot acute grave check breve
		sum prod coprod int oint iint iiint lim limsup liminf limits nolimits sin cos tan cot sec csc
		arcsin arccos arctan sinh cosh tanh log ln lg exp max min sup inf det dim ker gcd mod bmod pmod
		arg deg hom Pr infty partial nabla cdot cdots ldots dots vdots ddots times div pm mp ast star
		circ bullet leq le geq ge neq ne ll gg approx sim simeq cong equiv propto in notin ni subset
		subseteq supset supseteq cup cap setminus emptyset varnothing forall exists nexists neg land
		lor lnot oplus otimes odot to rightarrow leftarrow leftrightarrow Rightarrow Leftarrow
		Leftrightarrow mapsto longrightarrow longleftarrow uparrow downarrow quad qquad prime angle
		perp parallel mid langle rangle lfloor rfloor lceil rceil Re Im hbar ell wp aleph not
		displaystyle textstyle scriptstyle substack`) {
		m[c] = true
	}
	return m
}()

// safeTeX reports whether publisher TeX is safe to put between \( and \): every
// control sequence in it is a safe command or a safe escape, and no raw
// comment, math-shift or stray backslash is left.
func safeTeX(tex string, sym *Symbols) bool {
	var fromTable map[string]bool
	for _, m := range texControlRE.FindAllStringSubmatch(tex, -1) {
		name := m[1]
		if len(name) == 1 && !unicode.IsLetter(rune(name[0])) {
			if !strings.Contains(safeTeXEscapes, name) {
				return false
			}
			continue
		}
		if safeTeXCommands[name] {
			continue
		}
		if fromTable == nil {
			fromTable = map[string]bool{}
			for _, tbl := range []map[string]string{sym.TeX, sym.Accents, sym.MathVariants} {
				for _, cmd := range tbl {
					for _, n := range texCommandNameRE.FindAllStringSubmatch(cmd, -1) {
						fromTable[n[1]] = true
					}
				}
			}
		}
		if !fromTable[name] {
			return false
		}
	}
	return !strings.ContainsAny(texControlRE.ReplaceAllString(tex, ""), `\%$`)
}

// stripTeXDelims removes one pair of math delimiters ($$, \[ \], \( \), $) from
// the ends of publisher TeX.
func stripTeXDelims(s string) string {
	s = strings.TrimSpace(s)
	for _, d := range [][2]string{{"$$", "$$"}, {`\[`, `\]`}, {`\(`, `\)`}, {"$", "$"}} {
		if len(s) >= len(d[0])+len(d[1]) && strings.HasPrefix(s, d[0]) && strings.HasSuffix(s, d[1]) {
			s = s[len(d[0]) : len(s)-len(d[1])]
			break
		}
	}
	return strings.TrimSpace(s)
}

// normalizeHeading folds a heading for comparison: whitespace collapsed, a
// trailing colon removed, lower case.
func normalizeHeading(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToLower(strings.TrimRight(s, ": "))
}

// blockAttrs returns the attributes kept on a block element: the span of a
// table cell.
func blockAttrs(n *node) string {
	if n.name != "td" && n.name != "th" {
		return ""
	}
	var b strings.Builder
	for _, name := range []string{"colspan", "rowspan"} {
		if v := n.attrs[name]; v != "" && len(v) <= 4 && strings.Trim(v, "0123456789") == "" {
			b.WriteString(" " + name + `="` + v + `"`)
		}
	}
	return b.String()
}

// inlineAttrs returns the attributes kept on an inline element, and false when
// the element should be unwrapped (a link with no safe address).
func inlineAttrs(tag string, n *node) (string, bool) {
	var b strings.Builder
	switch tag {
	case "a":
		href := n.attrs["href"]
		if !safeURL(href) {
			return "", false
		}
		b.WriteString(` href="` + htmlAttrEscape(strings.TrimSpace(href)) + `"`)
		if t := n.attrs["title"]; t != "" {
			b.WriteString(` title="` + htmlAttrEscape(t) + `"`)
		}
	case "abbr", "acronym":
		if t := n.attrs["title"]; t != "" {
			b.WriteString(` title="` + htmlAttrEscape(t) + `"`)
		}
	}
	return b.String(), true
}

// safeURL accepts a non-empty address that is relative or uses http, https or
// mailto: the schemes the sanitizer keeps.
func safeURL(href string) bool {
	href = strings.TrimSpace(href)
	if href == "" || strings.IndexFunc(href, unicode.IsControl) >= 0 {
		return false
	}
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	switch u.Scheme { // url.Parse lower-cases the scheme
	case "", "http", "https", "mailto":
		return true
	}
	return false
}

// htmlEscape escapes the characters HTML treats specially in text.
func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// htmlAttrEscape escapes text for a double-quoted attribute value.
func htmlAttrEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
