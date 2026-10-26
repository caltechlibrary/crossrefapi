
# crossrefapi

This is a go package for working with the Crossref API. It is inspired by
the an excellent CrossRefAPI Python package listed in the Crossref API docs.
This package is meant to follow the "polite" guidelines for interacting 
with the public API at api.crossref.org.

## Go package example

```go
    appName := path.Base(os.Args[0])
    client, err := crossrefapi.NewCrossRefClient(appName, "jane.doe@library.example.edu")
    if err != nil {
        // handle error...
    }
    works, err := client.Works("10.1037/0003-066x.59.1.29")
   
    if err != nil {
        // handle error...
    }
    // continue processing your "works" result...
```

You can compare two copies of a "works" response and see what has changed.

```go

    appName := path.Base(os.Args[0])
    client, err := crossrefapi.NewCrossRefClient(appName, "jane.doe@library.example.edu")
    if err != nil {
        // handle error...
    }
    newWorks, err := client.Works("10.1037/0003-066x.59.1.29")
    if err != nil {
        // handle error...
    }
    // Fetch our previously saved works document.
    src, err := os.ReadFile("0003-066x.59.1.29.json")
    if err != nil {
        // handle error...
    }
    oldWorks := new(crossrefapi.Works)
    if err := json.Unmarshal(src, &oldWorks); err != nil {
        // handle error...
    }
    src, err = oldWorks.DiffAsJSON(newWorks)
    if err != nil {
        // handle error...
    }
    fmt.Println("Diff for 10.1037/0003-066x.59.1.29")
    fmt.Printf("\n%s\n", src)
```

## Text normalization (titles and abstracts)

From v1.1.0 the package turns Crossref's marked-up titles and JATS abstracts
into text and HTML an institutional repository can use: tags removed,
subscripts and formulas converted, layout whitespace tidied, and abstracts as
HTML that Invenio RDM's sanitizer keeps, with formulas as TeX in `\( ... \)`.

```go
    title, warnings := works.Message.FullTitle()
    // Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood: ...
    description, _ := works.Message.HTMLAbstract()
    // <p>The extraction of cosmological information from two-point statistics ...</p>
    others, _ := works.Message.AlternateTitles()
    // the record's other titles, if it has any
```

`crossrefapi.PlainText(s)` and `crossrefapi.HTMLAbstract(s)` do the same for a
string. `crossrefapi.NewNormalizer(path)` makes a `Normalizer` that uses your own
override of the symbol table. Everything that cannot be converted cleanly comes
back as a warning, so a program can hand those records to a person. See
[Text normalization](text_normalization.md) for the rules, the warning codes and
the override file.

## Command line example

```
    crossrefapi -mailto="jane.doe@library.example.edu" works "10.1037/0003-066x.59.1.29"
```

## Documentation

- [Installation](INSTALL.md)
- [User manual](user_manual.md)
- [Text normalization](text_normalization.md) - titles, subtitles and abstracts as text and HTML
- [crossrefapi(1)](crossrefapi.1.md) - the command line tool's manual page

## Reference

- [Crossref API Docs](https://api.crossref.org/)
- [Crossref Schemas](https://www.crossref.org/documentation/schema-library/)
- [CrossRefAPI](https://github.com/fabiobatalha/crossrefapi) - Python implementation
