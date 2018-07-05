#!/bin/bash

PKGASSETS=$(which pkgassets)
if [ "$PKGASSETS" = "" ]; then
    cat <<EOT >&2
You need to have pkgassets installed and in your path.
To install pkgassets try:

    go get -u github.com/caltechlibrary/pkgassets/... 

EOT
    exit 1
fi

function buildHelp() {
    PROG="$1"
    pkgassets -o "cmd/${PROG}/assets.go" -p main \
        -exclude="nav.md:topics.md" \
        -ext=".md" -strip-prefix="/" -strip-suffix=".md" \
        Help "docs"
    git add "cmd/${PROG}/assets.go"
}

# build Help assets 
buildHelp crossrefapi

# build Template assets

