// Command harness applies a repository's .github/settings.yml to GitHub.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// version names the build and is set at link time.
var version = "dev"

const usage = `usage: harness apply [--check] <owner/repo>
       harness --version
`

// repository matches owner/repo under GitHub's account and repository name characters.
var repository = regexp.MustCompile(`(?i)^[a-z\d]+(?:-[a-z\d]+)*/[-.\w]+$`)

// maxOwnerLength is the longest account name GitHub allows.
const maxOwnerLength = 39

// cli holds the streams and the apply step a command line runs against.
type cli struct {
	stdout, stderr io.Writer
	apply          func(repo string, check bool) error
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command line in args and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	return cli{stdout: stdout, stderr: stderr, apply: applySettings}.run(args)
}

func (c cli) run(args []string) int {
	flags := flag.NewFlagSet("harness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	showVersion := flags.Bool("version", false, "print the version")
	if err := flags.Parse(args); err != nil {
		return c.flagError(err)
	}
	if *showVersion {
		if _, err := fmt.Fprintf(c.stdout, "harness %s\n", version); err != nil {
			return 1
		}
		return 0
	}
	if flags.NArg() == 0 {
		return c.usageError(nil)
	}
	switch flags.Arg(0) {
	case "apply":
		return c.runApply(flags.Args()[1:])
	default:
		return c.usageError(fmt.Errorf("unknown command %q", flags.Arg(0)))
	}
}

// runApply reads the apply flags before or after the repository and applies.
func (c cli) runApply(args []string) int {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	check := flags.Bool("check", false, "print differences only")
	var repos []string
	for {
		if err := flags.Parse(args); err != nil {
			return c.flagError(err)
		}
		if flags.NArg() == 0 {
			break
		}
		repos = append(repos, flags.Arg(0))
		args = flags.Args()[1:]
	}
	if len(repos) != 1 {
		return c.usageError(errors.New("apply takes exactly one repository"))
	}
	repo := repos[0]
	if !validRepository(repo) {
		return c.usageError(fmt.Errorf("repository %q is not a valid owner/repo", repo))
	}
	if err := c.apply(repo, *check); err != nil {
		_, _ = fmt.Fprintf(c.stderr, "harness: %v\n", err)
		return 1
	}
	return 0
}

// validRepository accepts owner/repo names GitHub allows, other than the path names . and ..
func validRepository(name string) bool {
	if !repository.MatchString(name) {
		return false
	}
	owner, repo, _ := strings.Cut(name, "/")
	return len(owner) <= maxOwnerLength && repo != "." && repo != ".."
}

// applySettings brings repo in line with its settings, or with check reports the differences.
func applySettings(string, bool) error {
	return errors.New("apply is not implemented")
}

// flagError shows the usage on stdout for a help request and reports any other parse error.
func (c cli) flagError(err error) int {
	if errors.Is(err, flag.ErrHelp) {
		if _, werr := io.WriteString(c.stdout, usage); werr != nil {
			return 1
		}
		return 0
	}
	return c.usageError(err)
}

// usageError reports err with the usage on stderr, where a failed write has nowhere else to go.
func (c cli) usageError(err error) int {
	message := usage
	if err != nil {
		message = fmt.Sprintf("harness: %v\n%s", err, usage)
	}
	_, _ = io.WriteString(c.stderr, message)
	return 2
}
