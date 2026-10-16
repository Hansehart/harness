// Brings repositories in line with their declared settings.
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

// Process exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

const usage = `usage: harness apply [--check] <owner/repo>
       harness --version
`

var (
	errMissingCommand  = errors.New("missing command")
	errUnknownCommand  = errors.New("unknown command")
	errRepositoryCount = errors.New("apply takes exactly one repository")
	errNotImplemented  = errors.New("apply is not implemented")
)

// Holds the target repository and whether to only report differences.
type applyArgs struct {
	repo  repo.Name
	check bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// Executes the command line in args and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("harness")
	showVersion := flags.Bool("version", false, "print the version")

	err := flags.Parse(args)
	if err != nil {
		return failure(stdout, stderr, err)
	}

	// The version flag answers on its own, without a command.
	if *showVersion {
		return write(stdout, stderr, "harness "+version()+"\n")
	}

	return runCommand(flags.Args(), stdout, stderr)
}

// Runs the command named by the first argument with the rest as its input.
func runCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, errMissingCommand)
	}

	if args[0] != "apply" {
		err := fmt.Errorf("%w %q", errUnknownCommand, args[0])

		return usageError(stderr, err)
	}

	parsed, err := parseApply(args[1:])
	if err != nil {
		return failure(stdout, stderr, err)
	}

	// A failed apply is a runtime error, distinct from a usage error.
	err = applySettings(parsed)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "harness: %v\n", err)

		return exitError
	}

	return exitOK
}

// Accepts exactly one valid repository and the check switch in any order.
func parseApply(args []string) (applyArgs, error) {
	flags := newFlagSet("apply")
	check := flags.Bool("check", false, "print differences only")

	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return applyArgs{}, err
	}

	if len(positional) != 1 {
		return applyArgs{}, errRepositoryCount
	}

	name, err := repo.Parse(positional[0])
	if err != nil {
		return applyArgs{}, fmt.Errorf("apply: %w", err)
	}

	return applyArgs{repo: name, check: *check}, nil
}

// Reads flags anywhere and keeps arguments after the end-of-flags marker.
func parseInterspersed(flags *flag.FlagSet, args []string) ([]string, error) {
	var positional []string

	// Each round reads flags up to the next argument, then resumes behind it.
	for {
		err := flags.Parse(args)
		if err != nil {
			return nil, fmt.Errorf("%s flags: %w", flags.Name(), err)
		}

		// All input is read, or the rest follows the end-of-flags marker.
		rest := flags.Args()
		if len(rest) == 0 || stoppedAtMarker(args, rest) {
			return append(positional, rest...), nil
		}

		// The parse stopped at an argument; keep it and go on behind it.
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

// Reports whether the end-of-flags marker was the last input the parse read.
func stoppedAtMarker(args, rest []string) bool {
	consumed := len(args) - len(rest)

	return consumed > 0 && args[consumed-1] == "--"
}

// Reports the release tag or commit the binary was built from.
func version() string {
	info, ok := debug.ReadBuildInfo()
	// Go records the version in the binary; the fallback is Go's own marker.
	if !ok || info.Main.Version == "" {
		return "(devel)"
	}

	return info.Main.Version
}

// Creates a silent flag set whose errors the caller reports.
func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	return flags
}

// Brings a repository in line with its settings, or reports differences.
func applySettings(applyArgs) error {
	return errNotImplemented
}

// Prints the usage on stdout for help and reports other errors as misuse.
func failure(stdout, stderr io.Writer, err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return write(stdout, stderr, usage)
	}

	return usageError(stderr, err)
}

// Prints text on stdout, reporting a failed write on stderr.
func write(stdout, stderr io.Writer, text string) int {
	_, err := io.WriteString(stdout, text)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "harness: write output: %v\n", err)

		return exitError
	}

	return exitOK
}

// Prints an error and the usage on stderr, the last channel for output.
func usageError(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintf(stderr, "harness: %v\n%s", err, usage)

	return exitUsage
}
