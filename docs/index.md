
USAGE
=====

	crossrefapi [OPTIONS] types|works DOI

SYNOPSIS
--------


_crossrefapi_ can retrieve "types" and "works" from the CrossRef API


DESCRIPTION
-----------


_crossrefapi_ is a command line utility to retrieve "types" and "works" objects
from the CrossRef API. It follows the etiquette suggested at

```
    https://github.com/CrossRef/rest-api-doc#etiquette
```


OPTIONS
-------

Below are a set of options available.

```
    -generate-manpage    generate man page
    -generate-markdown   output documentation in Markdown
    -h, -help            display help
    -l, -license         display license
    -m, -mailto          set the mailto value for API access
    -v, -version         display app version
```


EXAMPLES
--------


Return the types of objects in CrossRef (e.g. journal articles)

```
    crossrefapi -mailto="jdoe@example.edu" types
```

Return the works for the doi "10.1037/0003-066x.59.1.29"

```
    crossrefapi -mailto="jdoe@example.edu" \
        works "10.1037/0003-066x.59.1.29"
```


crossrefapi v0.0.5
