//
// crossrefapi.go is a command line tool for access the CrossRef API given
// a specific DOI.
//
// Author R. S. Doiel, <rsdoiel@library.caltech.edu>
//
// Copyright (c) 2021, Caltech
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
//
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
	description = `
USAGE

    {appName}  [OPTIONS] DOI

SYNOPSIS

{appName} can retrieve "types" and "works" from the CrossRef API

DETAIL

{appName} is a command line utility to retrieve "types" and "works" objects
from the CrossRef API. It follows the etiquette suggested at

    https://github.com/CrossRef/rest-api-doc#etiquette

`
	examples = `
Return the types of objects in CrossRef (e.g. journal articles)

    {appName} -mailto="jdoe@example.edu" types

Return the works for the doi "10.1037/0003-066x.59.1.29"

    {appName} -mailto="jdoe@example.edu" \
        works "10.1037/0003-066x.59.1.29"
`

	license = `
{appName} {version}

Copyright (c) 2021, Caltech
All rights not granted herein are expressly reserved by Caltech.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`
	// Standard Options
	showHelp         bool
	showLicense      bool
	showVersion      bool

	// App Specific Options
	mailto string
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
	flagSet := flag.NewFlagSet(appName, flag.ContinueOnError)


	// Standard Options
	flagSet.BoolVar(&showHelp, "h", false, "display help")
	flagSet.BoolVar(&showHelp, "h,help", false, "display help")
	flagSet.BoolVar(&showLicense, "license", false, "display license")
	flagSet.BoolVar(&showVersion, "version", false, "display app version")

	// Application Options
	flagSet.StringVar(&mailto, "m", "", "set the mailto value for API access")
	flagSet.StringVar(&mailto, "m,mailto", "", "set the mailto value for API access")

	flagSet.Parse(os.Args)
	args := flagSet.Args()


	if showHelp {
		crossrefapi.DisplayUsage(os.Stdout, appName, flagSet, description, examples, license)
		os.Exit(0)
	}

	if showLicense {
		crossrefapi.DisplayLicense(os.Stdout, appName, license)
		os.Exit(0)
	}

	if showVersion {
		crossrefapi.DisplayVersion(os.Stdout, appName)
		os.Exit(0)
	}

	if len(args) < 1 {
		crossrefapi.DisplayUsage(os.Stderr, appName, flagSet, description, examples, license)
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
		obj, err := api.Works(doi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		src, err = json.MarshalIndent(obj, "", "    ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "USAGE: %s works DOI | %s types\n", appName, appName)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s\n", src)
	os.Exit(0)
}
