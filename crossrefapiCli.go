package crossrefapi

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

func DisplayLicense(out io.Writer, appName string, license string) {
	fmt.Fprintf(out, strings.ReplaceAll(strings.ReplaceAll(license, "{appName}", appName), "{version}", Version))
}

func DisplayVersion(out io.Writer, appName string) {
	fmt.Fprintf(out, "\n%s %s\n", appName, Version)
}

func DisplayUsage(out io.Writer, appName string, flagSet *flag.FlagSet, description string, examples string, license string) {
	// Convert {appName} and {version} in description
	if description != "" {
	fmt.Fprintf(out, strings.ReplaceAll(description, "{appName}", appName))
	}
	flagSet.SetOutput(out)
	flagSet.PrintDefaults()

	if examples != "" {
	fmt.Fprintf(out, strings.ReplaceAll(examples, "{appName}", appName))
	}
	if license != "" {
		DisplayLicense(out, appName, license)
	} 
	DisplayVersion(out, appName)
}

