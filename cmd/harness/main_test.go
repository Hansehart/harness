package main

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"
)

const (
	applyCmd    = "apply"
	checkFlag   = "--check"
	ownRepo     = "Hansehart/harness"
	someRepo    = "a/b"
	usageMark   = "usage:"
	versionFlag = "--version"
)

// Stands in for an output that refuses every write.
var errClosed = errors.New("closed")

// Exit code and output of one run.
type outcome struct {
	stdout, stderr string
	code           int
}

// Observes one run from the outside.
func runArgs(args ...string) outcome {
	var stdout, stderr bytes.Buffer

	code := run(args, &stdout, &stderr)

	return outcome{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func TestRunMisuse(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		nil,
		{"deploy"},
		{"--force"},
		{applyCmd},
		{applyCmd, someRepo, "c/d"},
	} {
		got := runArgs(args...)
		if got.code != exitUsage || got.stdout != "" {
			t.Errorf("run(%q) = %d, stdout %q; want %d and no stdout",
				args, got.code, got.stdout, exitUsage)
		}

		if !strings.Contains(got.stderr, usageMark) {
			t.Errorf("run(%q) stderr = %q, want the usage", args, got.stderr)
		}
	}
}

func TestRunNamesInvalidRepository(t *testing.T) {
	t.Parallel()

	got := runArgs(applyCmd, "harness")
	if got.code != exitUsage {
		t.Errorf("run = %d, want %d", got.code, exitUsage)
	}

	if !strings.Contains(got.stderr, `repository "harness"`) {
		t.Errorf("stderr = %q, want it to name the repository", got.stderr)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"help"},
		{"--help"},
		{"-h"},
		{applyCmd, "--help"},
	} {
		got := runArgs(args...)
		if got.code != exitOK || got.stderr != "" {
			t.Errorf("run(%q) = %d, stderr %q; want %d and no stderr",
				args, got.code, got.stderr, exitOK)
		}

		if !strings.Contains(got.stdout, usageMark) ||
			!strings.Contains(got.stdout, "harness help") {
			t.Errorf("run(%q) stdout = %q, want the usage", args, got.stdout)
		}
	}
}

func TestRunApplyNotImplemented(t *testing.T) {
	t.Parallel()

	got := runArgs(applyCmd, ownRepo)
	if got.code != exitError || got.stdout != "" {
		t.Errorf("run = %d, stdout %q; want %d and no stdout",
			got.code, got.stdout, exitError)
	}

	if !strings.Contains(got.stderr, errNotImplemented.Error()) {
		t.Errorf("stderr = %q, want %q", got.stderr, errNotImplemented)
	}
}

func TestParseApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantRepo  string
		args      []string
		wantCheck bool
	}{
		{
			name:      "repository only",
			wantRepo:  ownRepo,
			args:      []string{ownRepo},
			wantCheck: false,
		},
		{
			name:      "check before the repository",
			wantRepo:  someRepo,
			args:      []string{checkFlag, someRepo},
			wantCheck: true,
		},
		{
			name:      "check after the repository",
			wantRepo:  someRepo,
			args:      []string{someRepo, checkFlag},
			wantCheck: true,
		},
		{
			name:      "repository after the end-of-flags marker",
			wantRepo:  someRepo,
			args:      []string{"--", someRepo},
			wantCheck: false,
		},
	}

	for _, test := range tests {
		got, err := parseApply(test.args)
		if err != nil {
			t.Errorf("%s: parseApply(%q) error: %v", test.name, test.args, err)

			continue
		}

		if got.repo.String() != test.wantRepo || got.check != test.wantCheck {
			t.Errorf("%s: parseApply(%q) = (%s, %t), want (%s, %t)", test.name,
				test.args, got.repo, got.check, test.wantRepo, test.wantCheck)
		}
	}
}

func TestParseApplyInvalid(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		nil,
		{someRepo, "c/d"},
		{"--", someRepo, checkFlag},
		{"../.."},
		{"--force", someRepo},
	} {
		got, err := parseApply(args)
		if err == nil {
			t.Errorf("parseApply(%q) = %+v, want an error", args, got)
		}
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	// Builds name their version or mark themselves as development builds.
	want := regexp.MustCompile(`^harness (\(devel\)|v\d+\.\d+\.\d+\S*)\n$`)

	got := runArgs(versionFlag)
	if got.code != exitOK || got.stderr != "" {
		t.Fatalf("run = %d, stderr %q; want %d and no stderr",
			got.code, got.stderr, exitOK)
	}

	if !want.MatchString(got.stdout) {
		t.Errorf("stdout = %q, want it to match %s", got.stdout, want)
	}
}

// Refuses every write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errClosed
}

func TestRunVersionWriteFailure(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	got := run([]string{versionFlag}, failingWriter{}, &stderr)
	if got != exitError {
		t.Errorf("run with a failing stdout = %d, want %d", got, exitError)
	}

	if !strings.Contains(stderr.String(), errClosed.Error()) {
		t.Errorf("stderr = %q, want the write error", stderr.String())
	}
}
