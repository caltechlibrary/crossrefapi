---
title: "crossrefapi(1) user manual | version 1.0.6 358b959"
author: "R. S. Doiel"
pubDate: 2023-09-18
---

# NAME

crossrefapi

# SYNOPSIS

crossrefapi [OPTIONS] types|works DOI

# DESCRIPTION

crossrefapi can retrieve "types" and "works" from the CrossRef API. Is also
has the ability to compare the current "works" document with a JSON document
retrieved previously. The program uses the CrossRef REST API.
It follows the etiquette suggested at

~~~
	http://api.crossref.org/swagger-ui/index.html
~~~

# OPTIONS

-help
: display help

-license
: display license

-diff JSON_FILENAME
: compares the JSON_FILENAME with the current works retrieved and displays a diff between them as a JSON array where the first element is the old values
and the second element is the new values.

-mailto string
: set the mailto value for API access

-version
: display app version

# EXAMPLES

Return the types of objects in CrossRef (e.g. journal articles)

~~~
    crossrefapi -mailto="jdoe@example.edu" types
~~~

Return the works for the doi "10.1037/0003-066x.59.1.29"

~~~
    crossrefapi -mailto="jdoe@example.edu" \
        works "10.1037/0003-066x.59.1.29"
~~~

Compare a previously retrieved "works.json" with the current version.

~~~
crossrefapi -mailto="jdoe@example.edu" -diff works.json \
   works "10.1037/0003-066x.59.1.29"
~~~

