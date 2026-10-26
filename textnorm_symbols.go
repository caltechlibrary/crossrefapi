package crossrefapi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

//go:embed symbols.json
var defaultSymbolsJSON []byte

// Symbols is the mapping table the text normalizer reads. Each section maps
// one thing to another; a key that is absent from a section means "leave it
// as it is".
//
//   - Tokens maps the text of a MathML token (mi, mo, mn, mtext) to the text it
//     becomes. A value of "" removes the token (the invisible operators).
//   - Entities maps a bare character entity name ("nbsp", not "&nbsp;") to its
//     character, for entities the XML parser does not predefine.
//   - Superscripts and Subscripts map one character to its one-character
//     Unicode super or subscript form.
//   - TeX maps a character to the TeX command that produces it.
//   - Accents maps a MathML over-accent character to the TeX command that
//     puts it over its base (macron to \overline).
//   - MathVariants maps a MathML mathvariant value to the TeX font command
//     that applies it (double-struck to \mathbb).
//   - DroppedHeadings lists abstract headings that are dropped when they come
//     first ("Abstract"); matching rules belong to the abstract converter.
//
// The default table is embedded in the package (symbols.json). Use
// [DefaultSymbols] or [LoadSymbols] to get one; each call returns a new value.
//
// @example
//
//	s, err := crossrefapi.DefaultSymbols()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(s.Superscripts["2"], s.TeX["τ"]) // ² \tau
type Symbols struct {
	Tokens          map[string]string `json:"tokens"`
	Entities        map[string]string `json:"entities"`
	Superscripts    map[string]string `json:"superscripts"`
	Subscripts      map[string]string `json:"subscripts"`
	TeX             map[string]string `json:"tex"`
	Accents         map[string]string `json:"accents"`
	MathVariants    map[string]string `json:"mathvariants"`
	DroppedHeadings []string          `json:"dropped_headings"`
}

// symbolsOverride is the shape of an override file. DroppedHeadings is a
// pointer so an absent key (keep the default) can be told from an empty list
// (drop nothing).
type symbolsOverride struct {
	Tokens          map[string]string `json:"tokens"`
	Entities        map[string]string `json:"entities"`
	Superscripts    map[string]string `json:"superscripts"`
	Subscripts      map[string]string `json:"subscripts"`
	TeX             map[string]string `json:"tex"`
	Accents         map[string]string `json:"accents"`
	MathVariants    map[string]string `json:"mathvariants"`
	DroppedHeadings *[]string         `json:"dropped_headings"`
}

// DefaultSymbols returns the mapping table embedded in the package.
//
// @returns {*Symbols} a new, independent copy of the default table
// @returns {error} non-nil only if the embedded table is malformed, which the
// package's tests rule out
// @example
//
//	s, err := crossrefapi.DefaultSymbols()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(s.Superscripts["2"]) // ²
func DefaultSymbols() (*Symbols, error) {
	var s Symbols
	if err := decodeStrict(defaultSymbolsJSON, "embedded symbols.json", &s); err != nil {
		return nil, err
	}
	if err := validateSymbols(&s, "embedded symbols.json"); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadSymbols returns the default table with an override file applied. An
// override entry replaces the default entry with the same key, and a new key
// is added; entries the file does not name are kept. A "dropped_headings" list
// replaces the default list, so an empty list drops nothing; an absent key
// keeps the default.
//
// An empty overridePath means there is no override and is not an error. A path
// that does not exist, a file that is not valid JSON, an unknown section, or an
// invalid entry is an error that names the file.
//
// @param overridePath {string} path of a JSON override file, or "" for none
// @returns {*Symbols} the merged table, or nil when there is an error
// @returns {error} the problem, wrapping fs.ErrNotExist for a missing file
// @example
//
//	s, err := crossrefapi.LoadSymbols("my-symbols.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(s.TeX["τ"]) // \tau unless my-symbols.json says otherwise
func LoadSymbols(overridePath string) (*Symbols, error) {
	s, err := DefaultSymbols()
	if err != nil {
		return nil, err
	}
	if overridePath == "" {
		return s, nil
	}
	src, err := os.ReadFile(overridePath)
	if err != nil {
		return nil, fmt.Errorf("symbols override: %w", err)
	}
	var o symbolsOverride
	if err := decodeStrict(src, overridePath, &o); err != nil {
		return nil, err
	}
	merged := &Symbols{
		Tokens:          mergeMaps(s.Tokens, o.Tokens),
		Entities:        mergeMaps(s.Entities, o.Entities),
		Superscripts:    mergeMaps(s.Superscripts, o.Superscripts),
		Subscripts:      mergeMaps(s.Subscripts, o.Subscripts),
		TeX:             mergeMaps(s.TeX, o.TeX),
		Accents:         mergeMaps(s.Accents, o.Accents),
		MathVariants:    mergeMaps(s.MathVariants, o.MathVariants),
		DroppedHeadings: s.DroppedHeadings,
	}
	if o.DroppedHeadings != nil {
		merged.DroppedHeadings = append([]string{}, *o.DroppedHeadings...)
	}
	if err := validateSymbols(merged, overridePath); err != nil {
		return nil, err
	}
	return merged, nil
}

// mergeMaps returns a new map holding base with every entry of over added or
// replaced.
func mergeMaps(base, over map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(over))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		out[k] = v
	}
	return out
}

// decodeStrict decodes one JSON value from src into v, refusing unknown fields
// and anything after the value. Errors name source.
func decodeStrict(src []byte, source string, v any) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("%s: unexpected data after the JSON object", source)
	}
	return nil
}

// validateSymbols checks the shape rules of a table: no empty key, every
// super and subscript maps one character to one character, no TeX or entity
// value is empty, and no dropped heading is blank. Errors name source.
func validateSymbols(s *Symbols, source string) error {
	for section, m := range map[string]map[string]string{
		"tokens": s.Tokens, "entities": s.Entities, "superscripts": s.Superscripts,
		"subscripts": s.Subscripts, "tex": s.TeX, "accents": s.Accents, "mathvariants": s.MathVariants,
	} {
		for k, v := range m {
			if k == "" {
				return fmt.Errorf("%s: %s has an empty key", source, section)
			}
			switch section {
			case "superscripts", "subscripts":
				if utf8.RuneCountInString(k) != 1 || utf8.RuneCountInString(v) != 1 {
					return fmt.Errorf("%s: %s[%q] = %q: want one character mapped to one character", source, section, k, v)
				}
			case "tex", "entities":
				if v == "" {
					return fmt.Errorf("%s: %s[%q] has an empty value", source, section, k)
				}
			case "accents", "mathvariants":
				if !strings.HasPrefix(v, `\`) {
					return fmt.Errorf("%s: %s[%q] = %q: want a TeX command starting with a backslash", source, section, k, v)
				}
			}
		}
	}
	for _, h := range s.DroppedHeadings {
		if strings.TrimSpace(h) == "" {
			return fmt.Errorf("%s: dropped_headings has a blank entry", source)
		}
	}
	return nil
}
