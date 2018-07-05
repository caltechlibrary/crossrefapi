
# crossrefapi

This is a go package for working with the CrossRef API. It is inspired by
the an excellent CrossRefAPI Python package listed in the CrossRef API docs.
This package is meant to follow the "polite" guidelines for interacting 
with the public API at api.crossref.org.

## Go package example

```go
    client, err := crossrefapi.NewCrossRefClient("jane.doe@library.example.edu")
    if err != nil {
        // handle error...
    }
    works, err := client.Works("10.1037/0003-066x.59.1.29")
   
    if err != nil {
        // handle error...
    }
    // continue processing your "works" result...
```

## Command line example

```
    crossrefapi -mailto="jane.doe@library.example.edu" works "10.1037/0003-066x.59.1.29"
```

## Reference

+ [CrossRef API Docs](https://github.com/CrossRef/rest-api-doc)
+ [CrossRef Schemas](https://www.crossref.org/schema/)
+ [CrossRefAPI](https://github.com/fabiobatalha/crossrefapi) - Python implementation
