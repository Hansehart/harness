package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{name: "no arguments", args: nil, wantCode: 2, wantStderr: "usage:"},
		{name: "unknown command", args: []string{"deploy"}, wantCode: 2, wantStderr: "usage:"},
		{name: "unknown flag", args: []string{"--force"}, wantCode: 2, wantStderr: "usage:"},
		{name: "version", args: []string{"--version"}, wantCode: 0, wantStdout: "harness dev\n"},
		{name: "help", args: []string{"--help"}, wantCode: 0, wantStdout: "usage:"},
		{name: "short help", args: []string{"-h"}, wantCode: 0, wantStdout: "usage:"},
		{name: "apply help", args: []string{"apply", "--help"}, wantCode: 0, wantStdout: "usage:"},
		{name: "apply without repository", args: []string{"apply"}, wantCode: 2, wantStderr: "usage:"},
		{name: "apply with two repositories", args: []string{"apply", "a/b", "c/d"}, wantCode: 2, wantStderr: "usage:"},
		{name: "apply with a malformed repository", args: []string{"apply", "harness"}, wantCode: 2, wantStderr: "owner/repo"},
		{name: "apply until it is implemented", args: []string{"apply", "Hansehart/harness"}, wantCode: 1, wantStderr: "apply is not implemented"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tc.args, &stdout, &stderr)
			if got != tc.wantCode {
				t.Errorf("run(%q) = %d, want %d", tc.args, got, tc.wantCode)
			}
			checkStream(t, "stdout", tc.args, stdout.String(), tc.wantStdout)
			checkStream(t, "stderr", tc.args, stderr.String(), tc.wantStderr)
		})
	}
}

// checkStream requires an empty stream when want is empty and otherwise a stream containing want.
func checkStream(t *testing.T, stream string, args []string, got, want string) {
	t.Helper()
	if want == "" && got != "" {
		t.Errorf("run(%q) %s = %q, want it empty", args, stream, got)
	}
	if !strings.Contains(got, want) {
		t.Errorf("run(%q) %s = %q, want it to contain %q", args, stream, got, want)
	}
}

func TestApplyArguments(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantRepo  string
		wantCheck bool
	}{
		{name: "repository only", args: []string{"apply", "Hansehart/harness"}, wantRepo: "Hansehart/harness"},
		{name: "check before the repository", args: []string{"apply", "--check", "a/b"}, wantRepo: "a/b", wantCheck: true},
		{name: "check after the repository", args: []string{"apply", "a/b", "--check"}, wantRepo: "a/b", wantCheck: true},
		{name: "dotted and underscored name", args: []string{"apply", "octo-org/repo.name_1"}, wantRepo: "octo-org/repo.name_1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotRepo string
			var gotCheck bool
			var stdout, stderr bytes.Buffer
			c := cli{stdout: &stdout, stderr: &stderr, apply: func(repo string, check bool) error {
				gotRepo, gotCheck = repo, check
				return nil
			}}
			if code := c.run(tc.args); code != 0 {
				t.Fatalf("run(%q) = %d, want 0; stderr %q", tc.args, code, stderr.String())
			}
			if gotRepo != tc.wantRepo || gotCheck != tc.wantCheck {
				t.Errorf("run(%q) applied (%q, %t), want (%q, %t)", tc.args, gotRepo, gotCheck, tc.wantRepo, tc.wantCheck)
			}
		})
	}
}

func TestInvalidRepository(t *testing.T) {
	for _, name := range []string{
		" owner/repo",
		"owner/repo ",
		"../..",
		"a b/c",
		"owner/.",
		"owner/..",
		"-owner/repo",
		"owner-/repo",
		"own--er/repo",
		strings.Repeat("a", 40) + "/repo",
		strings.Repeat("a-", 20) + "a/repo",
		"owner/re?po",
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			c := cli{stdout: &stdout, stderr: &stderr, apply: func(string, bool) error {
				t.Errorf("apply ran for %q", name)
				return nil
			}}
			if code := c.run([]string{"apply", name}); code != 2 {
				t.Errorf("run(apply %q) = %d, want 2", name, code)
			}
			if !strings.Contains(stderr.String(), name) {
				t.Errorf("run(apply %q) stderr = %q, want it to name the repository", name, stderr.String())
			}
		})
	}
}

// failingWriter rejects every write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestRunVersionWriteFailure(t *testing.T) {
	var stderr bytes.Buffer
	if got := run([]string{"--version"}, failingWriter{}, &stderr); got != 1 {
		t.Errorf("run(--version) with a failing stdout = %d, want 1", got)
	}
}
