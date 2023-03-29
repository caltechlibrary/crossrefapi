
# crossrefapi

This is a go package for working with the CrossRef API. It is inspired by
the an excellent CrossRefAPI Python package listed in the CrossRef API docs.
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

## Command line example

```
    crossrefapi -mailto="jane.doe@library.example.edu" works "10.1037/0003-066x.59.1.29"
```

## Reference

- [CrossRef API Docs](https://github.com/CrossRef/rest-api-doc)
- [CrossRef Schemas](https://api.crossref.org/swagger-ui/index.html)
- [CrossRefAPI](https://github.com/fabiobatalha/crossrefapi) - Python implementation
