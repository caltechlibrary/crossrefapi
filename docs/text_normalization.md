# Text normalization for titles and abstracts

Crossref titles and abstracts are not plain text. A title can carry inline
markup (`<i>Euclid</i>`, `H<sub>2</sub>O`) and MathML formulas, indented with
newlines the publisher's layout added. An abstract is JATS XML. Pasted into an
institutional repository as they are, they show tags, run words together
(`with<mml:math>…</mml:math>current`) or lose their formulas. A repository
sanitizer such as Invenio RDM's removes MathML outright.

From v1.1.0 this package turns them into text a repository can use:

| You have | You call | You get |
|---|---|---|
| a title or subtitle string | `crossrefapi.PlainText(s)` | plain Unicode text |
| a work (`*Message`) | `msg.FullTitle()` | title and subtitle as plain text |
| a work | `msg.AlternateTitles()` | its other titles, as plain text |
| an abstract string, or a work | `crossrefapi.HTMLAbstract(s)`, `msg.HTMLAbstract()` | HTML the repository's sanitizer keeps |
| any of these, with your own symbol table | a `Normalizer` from `crossrefapi.NewNormalizer(path)` | the same, using your table |

Every function returns the text and a list of warnings. The text is always the
best available, and the warnings say what could not be converted cleanly, so a
program can send those records to a person.

```go
    client, _ := crossrefapi.NewCrossRefClient("myapp", "jane.doe@library.example.edu")
    works, err := client.Works("10.1051/0004-6361/202660273")
    if err != nil {
        // handle error...
    }
    title, warnings := works.Message.FullTitle()
    // Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood: ...
    description, _ := works.Message.HTMLAbstract()
    // <p>The extraction of cosmological information from two-point statistics ...</p>
```

Nothing here changes `Works.Message` or any existing function.

## Titles: PlainText

`PlainText` removes tags and keeps their text, decodes character entities
(`&lt;`, `&nbsp;`, `&#233;`), turns `<sub>` and `<sup>` into Unicode super and
subscripts, converts MathML formulas to Unicode, and tidies whitespace.

| Input | Output |
|---|---|
| `<i>Euclid</i>⏎   preparation` | `Euclid preparation` |
| `H⏎ <sub>2</sub>⏎ O:NH⏎ <sub>3</sub>⏎ :CO ices` | `H₂O:NH₃:CO ices` |
| `Ta1–⏎ <i>x</i>⏎ Hf⏎ <i>x</i>⏎ C0.5` | `Ta1–xHfxC0.5` |
| `with<mml:math>1/r²…</mml:math>current-current` | `with 1/r² current-current` |

Whitespace is the hard part, because a publisher's layout puts a newline and
indentation at every tag boundary and the original gap is lost. The rules are:

- Whitespace at the start or end, and runs of whitespace, collapse to nothing or
  a single space.
- A gap that contains a newline and touches a `<sub>`, a `<sup>` or a
  one-character italic is dropped, because a space never belongs there
  (`H₂O`, `¹²CO`, `Ta1–xHfxC0.5`).
- But if that gap also holds a typographic space the publisher wrote on purpose
  (a thin, en, em or no-break space), one space is kept: `(4 < M < 8)` stays
  `(4 < M < 8)`.
- A gap beside any other tag collapses to one space (`Euclid preparation`).
- Where a formula touches a word with no gap, a space is added
  (`with 1/r² current-current`).
- A space between a formula and a following hyphen goes (`τ-lepton`), and so
  does a layout newline between a formula and `, . ; ) ]` or after `( [`
  (`(N−1)`, `d=2, only`).

### Formulas in titles

MathML becomes Unicode, using the symbol table. Scripts become Unicode
superscripts and subscripts when every character has one (`x²`, `TaS₂`, `10⁻²`);
otherwise a linear form is written and a `math-linear` warning is returned
(`Z_N`, `x^(2q)`). A fraction is `a/b`, a square root `√x`, a root `³√x`; an
operand of more than one token is bracketed (`(x+1)/2`, `√(x+1)`). Any other
element (a table, an over-bar) is reduced to its text with a `math-unconverted`
warning naming the element.

```go
    title, warnings := crossrefapi.PlainText(
        `with <mml:math xmlns:mml="http://www.w3.org/1998/Math/MathML">` +
            `<mml:msub><mml:mi>Z</mml:mi><mml:mi>N</mml:mi></mml:msub></mml:math> symmetry`)
    // title is "with Z_N symmetry"; warnings holds "math-linear: Z_N"
```

A formula is recognized by the MathML namespace, not by its prefix: Crossref
uses `mml:` in some records and a different one-letter prefix for every formula
in others.

## Full title and alternate titles

`FullTitle` is the first title, then a colon and the first subtitle:
`Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood`.

- The subtitle is left out when it is empty, or when the title already contains
  it as whole words, ignoring case and punctuation (a book review whose title
  repeats its subtitle is not doubled).
- A title that already ends in `.`, `?`, `!` or `:` takes a space instead of a
  colon (`Reine Sprache? Privatbriefe…`).
- A record with no title (a journal issue) gives an empty string.

Some records have more than one title, usually in two languages, and Crossref
does not say which is which. `FullTitle` uses the first and does not guess.
When there are other distinct titles it returns a `multiple-titles` warning that
carries them, and `AlternateTitles` returns them as plain text, each once and in
order, so your program can decide which to use as the main title and where to
put the rest. A title that differs from the first only in capitalization is not
counted.

```go
    msg := &crossrefapi.Message{
        Title: []string{"X-chromosomale juvenile Retinoschisis", "X-linked juvenile retinoschisis"},
    }
    title, warnings := msg.FullTitle()  // "X-chromosomale juvenile Retinoschisis", one multiple-titles warning
    others, _ := msg.AlternateTitles()  // ["X-linked juvenile retinoschisis"]
```

## Abstracts: HTMLAbstract

`HTMLAbstract` returns HTML that Invenio RDM's sanitizer keeps. It was checked
against that sanitizer's tag list (`marshmallow_utils` 0.15.2, the version in
the RDM 13 and RDM 14 installations that were checked): every tag it writes is
on the list.

| Source | Output |
|---|---|
| `<jats:p>` | `<p>` |
| `<jats:title>` | `<p><b>…</b></p>`, except the first heading when it is "Abstract" (below) |
| `<jats:italic>`, `<jats:bold>` | `<i>`, `<b>` |
| `<jats:sub>`, `<jats:sup>`, and HTML `a abbr acronym b blockquote br code div em h1–h5 i li ol p pre s span strike strong sub sup u ul` and table tags (`table thead tbody tfoot tr th td`) | kept |
| `<jats:sec>`, `<jats:sc>`, any other element | the tag is removed, its text kept |
| `<jats:supplementary-material>`, `<jats:permissions>` (copyright statements), `<script>`, `<style>` | removed with their content |
| MathML | TeX inside `\( … \)` |
| `<jats:tex-math>` | ignored, unless a formula has no MathML at all |

**The first heading.** About six abstracts in ten begin with the heading
"Abstract", which only adds noise to a repository's description field. The first
heading is dropped when it appears in the symbol table's `dropped_headings`
list (by default just `Abstract`, ignoring case and a trailing colon). Only the
first is dropped, and only if nothing precedes it. Every other heading, such as
Background or Methods, is kept as a bold paragraph.

**Attributes.** All attributes are removed except `href` and `title` on a link,
`title` on `abbr` and `acronym`, and `colspan` and `rowspan` on a table cell. A
link keeps its address only if it is `http`, `https`, `mailto` or relative;
otherwise the link is unwrapped and its text kept.

**Formulas** are converted from MathML to TeX and written between `\(` and `\)`,
which is MathJax's default inline delimiter:

```go
    html, _ := crossrefapi.HTMLAbstract(
        `<jats:p>The <jats:italic>Euclid</jats:italic> mission, with ` +
            `<mml:math xmlns:mml="http://www.w3.org/1998/Math/MathML"><mml:mi>J</mml:mi><mml:mo>/</mml:mo>` +
            `<mml:mi>ψ</mml:mi></mml:math> data.</jats:p>`)
    // <p>The <i>Euclid</i> mission, with \(J/\psi\) data.</p>
```

The TeX needs a renderer on the page that reads those delimiters. RDM keeps the
text; whether and how it is drawn depends on the instance's MathJax
configuration, so check a test record before relying on it. Greek letters and
operators become TeX commands from the symbol table, a multi-letter identifier
becomes `\mathrm{…}`, an over-bar `\overline{…}`, a double-struck `Z`
`\mathbb{Z}`, and special characters (`% & # _ { }`) are escaped. Where the
publisher also supplied TeX, the MathML is converted anyway so the output is
consistent; the publisher's TeX is used only for a formula that has no MathML.

Whitespace follows the same rules as titles, and there is none between blocks.
The result is HTML: do not pass it to `PlainText`.

## Warnings

| Code | Meaning | `Fragment` |
|---|---|---|
| `parse-fallback` | the fragment is not well-formed (a raw `<`, a stray end tag); tags were stripped and entities decoded, and the whitespace rules were not applied | the whole input |
| `script-unmapped` | a `<sub>` or `<sup>` has a character with no Unicode form; its text was kept as it was | the element, such as `<sub>q</sub>` |
| `math-linear` | a formula script or root could not be shown in Unicode and is written in a linear form | the linear text, such as `Z_N` |
| `math-unconverted` | a MathML element has no conversion; the formula was reduced to its text | the element name |
| `multiple-titles` | the record has other distinct titles | the other titles, joined with a vertical bar and spaces |

A `Warning` prints as `code: fragment`.

## The symbol table and the override file

All the mappings live in one table, embedded in the package (`symbols.json`),
so a program does not depend on any file at run time. You can add to it or
change it with an override file, and use it through a `Normalizer`:

```go
    n, err := crossrefapi.NewNormalizer("my-symbols.json")
    if err != nil {
        // handle error...
    }
    title, warnings := n.FullTitle(works.Message)
    description, _ := n.AbstractHTML(works.Message)
```

An empty path means the embedded default. A `Normalizer` never changes after it
is built and holds no global state, so one value can be used from many
goroutines, and two with different files do not affect each other or the
package-level functions.

The file is JSON. Each section maps keys to values; an entry in your file
replaces the default entry with the same key and a new key is added, and
anything you do not mention is kept. The exception is `dropped_headings`, a
list that replaces the default list as a whole, so `[]` drops no heading.

| Section | Maps | Example |
|---|---|---|
| `tokens` | the text of a MathML token to what it becomes, in titles and abstracts (`""` removes it; the invisible operators are removed by default) | `"×": "x"` |
| `entities` | a character entity name (no `&` or `;`) to its character | `"eacute": "é"` |
| `superscripts`, `subscripts` | one character to its one-character Unicode form | `"q": "ᵠ"` |
| `tex` | a character to its TeX command | `"ℵ": "\\aleph"` |
| `accents` | an over-accent character to the TeX command that puts it over its base | `"¯": "\\overline"` |
| `mathvariants` | a MathML `mathvariant` to its TeX font command | `"double-struck": "\\mathbb"` |
| `dropped_headings` | the list of abstract headings dropped when they come first | `["Abstract", "Zusammenfassung"]` |

```json
{
  "tex": { "ℵ": "\\aleph" },
  "dropped_headings": ["Abstract", "Zusammenfassung"]
}
```

`NewNormalizer` returns an error that names the file when it does not exist,
is not valid JSON, has a section it does not know (a misspelt name is not
silently ignored), has trailing data, or has an invalid entry (a superscript or
subscript key or value that is not a single character, an empty key, an empty
TeX or entity value, or an accent or font command that does not start with a
backslash).

## What to know before relying on it

- **Whitespace is a judgement.** The original spacing around markup is not in
  the data, so the rules above are heuristics, chosen on real records. A
  one-letter italic next to a plain word with only layout whitespace between
  them is joined to it (`axb`), which is wrong for a rare record.
- **Only what real records use is converted.** MathML tokens, scripts,
  prescripts, fractions, roots, accents and limits are; tables, `semantics` and
  `mmultiscripts` are reduced to their text with a warning. In an abstract,
  JATS lists, links inside `ext-link`, and cross-references are unwrapped,
  keeping their text.
- **Titles ignore `mathvariant`.** A double-struck `Z` is `Z` in a title and
  `\mathbb{Z}` in an abstract.
- **Table captions and `<pre>` layout.** `caption`, `col` and `colgroup` are not
  on the sanitizer's list and are unwrapped; whitespace inside `<pre>` is
  collapsed like any other.
- **The repository may change characters.** RDM's sanitizer also normalizes
  text (for example curly quotes to straight ones); that happens after this
  package and is outside it.
- **No language guessing.** Which of several titles to lead with is your
  decision; `AlternateTitles` gives you the rest.
