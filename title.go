package crossrefapi

import (
	"strings"
	"unicode"
)

// FullTitle returns a record's title as plain text: the first title, then a
// colon and the first subtitle. The subtitle is left out when it is empty or the
// title already contains it, comparing whole words without regard to case or
// punctuation; a title that already ends in . ? ! or : takes a space instead of
// a colon. Only the first title is used; when the record has other distinct
// titles a multiple-titles warning carries them, and AlternateTitles returns
// them. A nil record, or one with no title, gives an empty string.
//
// @param msg {*Message} a CrossRef work, which may be nil
// @returns {string} the full title as plain text
// @returns {[]Warning} conversion warnings, and a multiple-titles warning when there are other titles
// @example
//
//	title, warnings := n.FullTitle(&msg)
//	fmt.Println(title) // Euclid preparation: CII. Non-Gaussianity of ...
func (n *Normalizer) FullTitle(msg *Message) (string, []Warning) {
	if msg == nil || len(msg.Title) == 0 {
		return "", nil
	}
	sym := n.table()
	title, ws := plainText(msg.Title[0], sym)
	if alts, _ := n.AlternateTitles(msg); len(alts) > 0 {
		ws = append(ws, Warning{Code: WarnMultipleTitles, Fragment: strings.Join(alts, " | ")})
	}
	if len(msg.SubTitle) == 0 {
		return title, ws
	}
	sub, subWarnings := plainText(msg.SubTitle[0], sym)
	ws = append(ws, subWarnings...)
	switch {
	case sub == "":
		return title, ws
	case title == "":
		return sub, ws
	case containsWords(title, sub):
		return title, ws
	}
	return title + titleJoiner(title) + sub, ws
}

// AlternateTitles returns the titles after the first as plain text, each once,
// in order. A title is left out when it is empty or equals, ignoring case, the
// first title or one already returned, so the same title in two letter cases
// counts once. No language is guessed: which title a record should lead with is
// the caller's decision. A nil record, or one with fewer than two titles, gives
// nil.
//
// @param msg {*Message} a CrossRef work, which may be nil
// @returns {[]string} the other titles, none repeated; nil when there are none
// @returns {[]Warning} conversion warnings for the titles returned
// @example
//
//	others, _ := n.AlternateTitles(&msg)
//	for _, t := range others {
//		fmt.Println(t)
//	}
func (n *Normalizer) AlternateTitles(msg *Message) ([]string, []Warning) {
	if msg == nil || len(msg.Title) < 2 {
		return nil, nil
	}
	sym := n.table()
	first, _ := plainText(msg.Title[0], sym)
	seen := map[string]bool{strings.ToLower(first): true}
	var out []string
	var ws []Warning
	for _, raw := range msg.Title[1:] {
		t, w := plainText(raw, sym)
		if t == "" || seen[strings.ToLower(t)] {
			continue
		}
		seen[strings.ToLower(t)] = true
		out = append(out, t)
		ws = append(ws, w...)
	}
	return out, ws
}

// FullTitle returns the record's title as plain text with the default symbol
// table; see (*Normalizer).FullTitle.
//
// @returns {string} the full title as plain text
// @returns {[]Warning} conversion warnings, and a multiple-titles warning when there are other titles
// @example
//
//	title, warnings := msg.FullTitle()
func (msg *Message) FullTitle() (string, []Warning) {
	return defaultNormalizer.FullTitle(msg)
}

// AlternateTitles returns the titles after the first as plain text with the
// default symbol table; see (*Normalizer).AlternateTitles.
//
// @returns {[]string} the other titles, none repeated; nil when there are none
// @returns {[]Warning} conversion warnings for the titles returned
// @example
//
//	others, _ := msg.AlternateTitles()
func (msg *Message) AlternateTitles() ([]string, []Warning) {
	return defaultNormalizer.AlternateTitles(msg)
}

// HTMLAbstract returns the record's abstract as sanitizer-safe HTML with the
// default symbol table; see the package-level HTMLAbstract for the rules.
//
// @returns {string} HTML, empty when the record has no abstract
// @returns {[]Warning} what could not be converted cleanly; empty when nothing
// @example
//
//	html, warnings := msg.HTMLAbstract()
func (msg *Message) HTMLAbstract() (string, []Warning) {
	return defaultNormalizer.AbstractHTML(msg)
}

// titleJoiner returns what goes between a title and its subtitle.
func titleJoiner(title string) string {
	if strings.ContainsRune(".?!:", lastRune(strings.TrimSpace(title))) {
		return " "
	}
	return ": "
}

func lastRune(s string) rune {
	r := []rune(s)
	if len(r) == 0 {
		return 0
	}
	return r[len(r)-1]
}

// containsWords reports whether the words of sub appear in order, as whole
// words, in title, ignoring case and punctuation.
func containsWords(title, sub string) bool {
	t, s := foldWords(title), foldWords(sub)
	return strings.Contains(" "+t+" ", " "+s+" ")
}

// foldWords lower-cases a string and reduces it to its letters and numbers,
// separated by single spaces.
func foldWords(s string) string {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return strings.Join(words, " ")
}
