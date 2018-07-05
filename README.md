
# crossrefapi

This is a go package for working with the CrossRef API. It is inspired by
the an excellent CrossRefAPI Python package listed in the CrossRef API docs.
This package is meant to follow the "polite" guidelines for interacting 
with the public API at api.crossref.org.

```go
    client, err := crossrefapi.NewCrossRefClient("jane.doe@library.example.edu")
    if err != nil {
        // handle error...
    }
    works, err := client.Works("10.1037/0003-066x.59.1.29")
   
    if err != nil {
        // handle error...
    }
    // continue processing your "work" result...
```

## Reference

+ [CrossRef API Docs](https://github.com/CrossRef/rest-api-doc)
+ [CrossRef Schemas](https://www.crossref.org/schema/)
+ [CrossRefAPI](https://github.com/fabiobatalha/crossrefapi) - Python implementation
