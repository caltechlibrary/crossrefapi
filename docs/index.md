
# USAGE

	crossrefapi [OPTIONS] types|works DOI

## SYNOPSIS


crossrefapi is a command line utility to retrieve "types" and "works" objects
from the CrossRef API. It follows the etiquette suggested at
	
	https://github.com/CrossRef/rest-api-doc#etiquette

EXAMPLES

Return the types of objects in CrossRef (e.g. journal articles)

	crossrefapi -mailto="jane.doe@example.edu" types

Return the works for the doi "10.1000/xyz123"

	crossrefapi -mailto="jane.doe@example.edu" works "10.1000/xyz123"



## OPTIONS

```
    -generate-markdown-docs   output documentation in Markdown
    -h, -help                 display help
    -l, -license              display license
    -m, -mailto               set the mailto value for API access
    -v, -version              display app version
```


crossrefapi v0.0.1
