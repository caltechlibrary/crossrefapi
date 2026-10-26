package crossrefapi

// Normalizer converts CrossRef titles and abstracts with one symbol table: the
// embedded default, or the default with an override file applied. A Normalizer
// never changes after it is made, so one value can be shared between
// goroutines; it holds no global state, so two Normalizers with different
// override files do not affect each other or the package's own functions.
//
// The zero value, and a nil *Normalizer, use the embedded default table. The
// package-level PlainText and HTMLAbstract functions and the (*Message) methods
// FullTitle, AlternateTitles and HTMLAbstract are the same conversions with that
// default; use a Normalizer when you need an override file (crossrefapi/DR-0001
// decision 4).
type Normalizer struct {
	sym *Symbols
}

// NewNormalizer returns a Normalizer whose table is the embedded default with an
// override file applied, as LoadSymbols does. An override entry replaces the
// default entry with the same key, a new key is added, and a "dropped_headings"
// list replaces the default list.
//
// @param overridePath {string} path of a JSON override file, or "" for the default table
// @returns {*Normalizer} a Normalizer, or nil when there is an error
// @returns {error} a missing file (wrapping fs.ErrNotExist), invalid JSON, an unknown section or an invalid entry, each naming the file
// @example
//
//	n, err := crossrefapi.NewNormalizer("my-symbols.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	title, warnings := n.FullTitle(msg)
func NewNormalizer(overridePath string) (*Normalizer, error) {
	sym, err := LoadSymbols(overridePath)
	if err != nil {
		return nil, err
	}
	return &Normalizer{sym: sym}, nil
}

// defaultNormalizer backs the (*Message) methods.
var defaultNormalizer = &Normalizer{}

// table returns the symbol table to convert with.
func (n *Normalizer) table() *Symbols {
	if n == nil || n.sym == nil {
		return defaultTable()
	}
	return n.sym
}

// PlainText converts a title fragment to plain Unicode text with this
// Normalizer's table; see the package-level PlainText for the rules.
//
// @param s {string} a CrossRef title, subtitle or other inline fragment
// @returns {string} the plain text
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	text, _ := n.PlainText("H<sub>2</sub>O")
//	fmt.Println(text) // H₂O
func (n *Normalizer) PlainText(s string) (string, []Warning) {
	return plainText(s, n.table())
}

// HTMLAbstract converts an abstract string to sanitizer-safe HTML with this
// Normalizer's table; see the package-level HTMLAbstract for the rules.
//
// @param s {string} a CrossRef abstract
// @returns {string} HTML, empty when the abstract has no text
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	html, _ := n.HTMLAbstract("<jats:p>One</jats:p>")
//	fmt.Println(html) // <p>One</p>
func (n *Normalizer) HTMLAbstract(s string) (string, []Warning) {
	return htmlAbstract(s, n.table())
}

// AbstractHTML converts a record's abstract to HTML, as HTMLAbstract does for a
// string. A nil record, or one with no abstract, gives an empty string.
//
// @param msg {*Message} a CrossRef work, which may be nil
// @returns {string} HTML, empty when there is no abstract
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	html, warnings := n.AbstractHTML(&msg)
func (n *Normalizer) AbstractHTML(msg *Message) (string, []Warning) {
	if msg == nil {
		return "", nil
	}
	return n.HTMLAbstract(msg.Abstract)
}
