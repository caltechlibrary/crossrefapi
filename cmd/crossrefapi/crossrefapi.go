// crossrefapi.go is a command line tool for access the CrossRef API given
// a specific DOI.
//
// Author R. S. Doiel, <rsdoiel@library.caltech.edu>
//
// Copyright (c) 2023, Caltech
// All rights not granted herein are expressly reserved by Caltech.
//
// Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/crossrefapi"
)

var (
	helpText = `---
title: "{app_name}(1) user manual | version {version} {release_hash}"
author: "R. S. Doiel"
pubDate: {release_date}
---

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] types|works DOI

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

`

	// Standard Options
	showHelp    bool
	showLicense bool
	showVersion bool

	// App Specific Options
	diffFName string
	mailto    string
)

func pop(args []string) (string, []string) {
	var (
		arg string
		l   int
	)
	l = len(args)
	switch {
	case l > 1:
		arg = args[0]
		args = args[1:]
	case l == 1:
		arg = args[0]
		args = []string{}
	default:
		return "", []string{}
	}
	return arg, args
}

func main() {
	appName := path.Base(os.Args[0])
	// NOTE: This is the date that version.go was generated.
	version := crossrefapi.Version
    releaseDate := crossrefapi.ReleaseDate
    releaseHash := crossrefapi.ReleaseHash
    fmtHelp := crossrefapi.FmtHelp

	flagSet := flag.NewFlagSet(appName, flag.ContinueOnError)

	// Standard Options
	flagSet.BoolVar(&showHelp, "help", false, "display help")
	flagSet.BoolVar(&showLicense, "license", false, "display license")
	flagSet.BoolVar(&showVersion, "version", false, "display app version")

	// Application Options
	flagSet.StringVar(&diffFName, "diff", "", "compare the current works results with JSON document")
	flagSet.StringVar(&mailto, "mailto", "", "set the mailto value for API access")

	flagSet.Parse(os.Args[1:])
	args := flagSet.Args()

	if showHelp {
		fmt.Fprint(os.Stdout, fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}

	if showLicense {
		fmt.Fprintf(os.Stdout, "%s\n", crossrefapi.LicenseText)
		os.Exit(0)
	}

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}

	if len(args) < 1 {
		fmt.Fprint(os.Stderr, fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(1)
	}

	api, err := crossrefapi.NewCrossRefClient(appName, mailto)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	var (
		src     []byte
		apiPath string
		doi     string
	)
	apiPath, args = pop(args)
	doi, args = pop(args)
	if apiPath == "" {
		fmt.Fprintf(os.Stderr, "USAGE: %s works DOI | %s types\n", appName, appName)
		os.Exit(1)
	}
	switch strings.ToLower(apiPath) {
	case "types":
		obj, err := api.Types()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		src, err = json.MarshalIndent(obj, "", "   ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	case "works":
		nWork, err := api.Works(doi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if nWork == nil {
			fmt.Fprintf(os.Stderr, "Missing works JSON from request")
			os.Exit(1)
		}
		if diffFName == "" {
			src, err = json.MarshalIndent(nWork, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		} else {
			src, err = os.ReadFile(diffFName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			oWork := new(crossrefapi.Works)
			if err := json.Unmarshal(src, &oWork); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			src, err = oWork.DiffAsJSON(nWork)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "USAGE: %s works DOI | %s types\n", appName, appName)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s\n", src)
	os.Exit(0)
}
