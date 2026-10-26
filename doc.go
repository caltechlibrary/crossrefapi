// Package crossrefapi is a client for the public Crossref REST API, with types
// for the works it returns and a text layer that turns Crossref's titles and
// abstracts into text and HTML an institutional repository can use.
//
// # Client
//
// NewCrossRefClient makes a client that identifies the application and a contact
// address, as Crossref's polite guidelines ask. Its Works method retrieves a work
// by DOI; Works.DiffAsJSON compares two copies of a work.
//
// # Text normalization
//
// Crossref titles carry inline markup (<i>Euclid</i>, H<sub>2</sub>O, MathML
// formulas) and abstracts are JATS XML. These functions and one type convert them:
//
//   - PlainText turns a title fragment into plain Unicode text.
//   - (*Message).FullTitle joins a title and its subtitle as plain text, and
//     (*Message).AlternateTitles returns the record's other titles.
//   - HTMLAbstract, and (*Message).HTMLAbstract, turn an abstract into HTML that
//     RDM's sanitizer keeps, with formulas as TeX in \( ... \).
//   - Normalizer does the same with a symbol table you supply as an override
//     file (NewNormalizer).
//
// Anything that cannot be converted cleanly comes back as a Warning beside the
// best text available. The reference is docs/text_normalization.md.
package crossrefapi
