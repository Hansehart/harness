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
       harness help
       harness --version
`

var (
	errMissingCommand  = errors.New("missing command")
	errNotImplemented  = errors.New("apply is not implemented")
	errRepositoryCount = errors.New("apply takes exactly one repository")
	errUnknownCommand  = errors.New("unknown command")
)

// Holds the target repository and whether to only report differences.
type applyArgs struct {
	repo  repo.Name
	check bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// Turns a command line into an exit code.
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

// Dispatches to the requested command.
func runCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, errMissingCommand)
	}

	if args[0] == "help" {
		return write(stdout, stderr, usage)
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
		// A failed report has nowhere left to go.
		_, _ = fmt.Fprintf(stderr, "harness: %v\n", err)

		return exitError
	}

	return exitOK
}

// Reads an apply request and refuses anything ambiguous.
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

// Lets options appear anywhere on the command line.
func parseInterspersed(flags *flag.FlagSet, args []string) ([]string, error) {
	var positional []string

	// Option parsing stops at each argument, so it resumes behind it.
	for {
		err := flags.Parse(args)
		if err != nil {
			return nil, fmt.Errorf("%s flags: %w", flags.Name(), err)
		}

		// Nothing is left to scan for options.
		rest := flags.Args()
		if len(rest) == 0 || stoppedAtMarker(args, rest) {
			return append(positional, rest...), nil
		}

		// Keep the argument and look for options after it.
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

// Tells whether option scanning was ended explicitly.
func stoppedAtMarker(args, rest []string) bool {
	consumed := len(args) - len(rest)

	return consumed > 0 && args[consumed-1] == "--"
}

// Identifies the build.
func version() string {
	info, ok := debug.ReadBuildInfo()
	// Builds without a recorded version count as development builds.
	if !ok || info.Main.Version == "" {
		return "(devel)"
	}

	return info.Main.Version
}

// Leaves all error reporting to the caller.
func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	return flags
}

// Brings a repository in line with its settings, or reports differences.
func applySettings(applyArgs) error {
	return errNotImplemented
}

// Treats a help request as success and anything else as misuse.
func failure(stdout, stderr io.Writer, err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return write(stdout, stderr, usage)
	}

	return usageError(stderr, err)
}

// Delivers output and makes a lost write visible.
func write(stdout, stderr io.Writer, text string) int {
	_, err := io.WriteString(stdout, text)
	if err != nil {
		// A failed report has nowhere left to go.
		_, _ = fmt.Fprintf(stderr, "harness: write output: %v\n", err)

		return exitError
	}

	return exitOK
}

// Explains the misuse and how to call the command.
func usageError(stderr io.Writer, err error) int {
	// A failed report has nowhere left to go.
	_, _ = fmt.Fprintf(stderr, "harness: %v\n%s", err, usage)

	return exitUsage
}
