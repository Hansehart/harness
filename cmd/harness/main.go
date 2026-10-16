// Command harness brings repositories in line with their declared settings.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/Hansehart/harness/internal/repo"
)

const usage = `usage: harness apply [--check] <owner/repo>
       harness --version
`

// applyArgs holds the target repository and whether to only report differences.
type applyArgs struct {
	repo  repo.Name
	check bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command line in args and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("harness")
	showVersion := flags.Bool("version", false, "print the version")
	if err := flags.Parse(args); err != nil {
		return failure(stdout, stderr, err)
	}
	if *showVersion {
		return write(stdout, "harness "+version()+"\n")
	}
	if flags.NArg() == 0 {
		return usageError(stderr, errors.New("missing command"))
	}
	if flags.Arg(0) != "apply" {
		return usageError(stderr, fmt.Errorf("unknown command %q", flags.Arg(0)))
	}
	parsed, err := parseApply(flags.Args()[1:])
	if err != nil {
		return failure(stdout, stderr, err)
	}
	if err := applySettings(parsed.repo, parsed.check); err != nil {
		_, _ = fmt.Fprintf(stderr, "harness: %v\n", err)
		return 1
	}
	return 0
}

// parseApply accepts exactly one valid repository and the check switch in any order.
func parseApply(args []string) (applyArgs, error) {
	flags := newFlagSet("apply")
	check := flags.Bool("check", false, "print differences only")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return applyArgs{}, err
	}
	if len(positional) != 1 {
		return applyArgs{}, errors.New("apply takes exactly one repository")
	}
	name, err := repo.Parse(positional[0])
	if err != nil {
		return applyArgs{}, err
	}
	return applyArgs{repo: name, check: *check}, nil
}

// parseInterspersed accepts flags in any position and takes every argument after "--" verbatim.
func parseInterspersed(flags *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := flags.Parse(args); err != nil {
			return nil, err
		}
		rest := flags.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		if consumed := len(args) - len(rest); consumed > 0 && args[consumed-1] == "--" {
			return append(positional, rest...), nil
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

// version reports the release tag or commit the binary was built from.
func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

// newFlagSet creates a silent flag set whose errors the caller reports.
func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

// applySettings brings name in line with its settings, or with check reports the differences.
func applySettings(repo.Name, bool) error {
	return errors.New("apply is not implemented")
}

// failure prints the usage on stdout for help and reports other errors as usage errors.
func failure(stdout, stderr io.Writer, err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return write(stdout, usage)
	}
	return usageError(stderr, err)
}

// write prints s on stdout and returns exit code 0, or 1 when the write fails.
func write(stdout io.Writer, s string) int {
	if _, err := io.WriteString(stdout, s); err != nil {
		return 1
	}
	return 0
}

// usageError prints err and the usage on stderr, where a failed write has nowhere else to go.
func usageError(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintf(stderr, "harness: %v\n%s", err, usage)
	return 2
}
