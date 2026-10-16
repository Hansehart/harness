// Command harness applies a repository's .github/settings.yml to GitHub.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// version names the build and is set at link time.
var version = "dev"

const usage = `usage: harness apply [--check] <owner/repo>
       harness --version
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command line in args and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("harness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	showVersion := flags.Bool("version", false, "print the version")
	if err := flags.Parse(args); err != nil {
		return usageError(stderr, err)
	}
	if *showVersion {
		if _, err := fmt.Fprintf(stdout, "harness %s\n", version); err != nil {
			return 1
		}
		return 0
	}
	if flags.NArg() == 0 {
		return usageError(stderr, nil)
	}
	switch flags.Arg(0) {
	case "apply":
		return apply(flags.Args()[1:], stderr)
	default:
		return usageError(stderr, fmt.Errorf("unknown command %q", flags.Arg(0)))
	}
}

// apply validates the apply arguments.
func apply(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Bool("check", false, "print differences only")
	if err := flags.Parse(args); err != nil {
		return usageError(stderr, err)
	}
	if flags.NArg() != 1 {
		return usageError(stderr, errors.New("apply takes exactly one repository"))
	}
	if err := checkRepository(flags.Arg(0)); err != nil {
		return usageError(stderr, err)
	}
	return 0
}

// checkRepository accepts a repository written as owner/repo.
func checkRepository(name string) error {
	owner, repo, found := strings.Cut(name, "/")
	if !found || owner == "" || repo == "" || strings.Contains(repo, "/") {
		return fmt.Errorf("repository %q is not written as owner/repo", name)
	}
	return nil
}

// usageError reports err and the usage on stderr, the last channel left, and returns the usage exit code.
func usageError(stderr io.Writer, err error) int {
	message := usage
	if err != nil {
		message = fmt.Sprintf("harness: %v\n%s", err, usage)
	}
	_, _ = io.WriteString(stderr, message)
	return 2
}
