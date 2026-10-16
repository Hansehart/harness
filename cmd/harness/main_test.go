package main

import (
	"bytes"
	"errors"
	"regexp"
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

func TestParseApply(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantRepo  string
		wantCheck bool
	}{
		{name: "repository only", args: []string{"Hansehart/harness"}, wantRepo: "Hansehart/harness"},
		{name: "check before the repository", args: []string{"--check", "a/b"}, wantRepo: "a/b", wantCheck: true},
		{name: "check after the repository", args: []string{"a/b", "--check"}, wantRepo: "a/b", wantCheck: true},
		{name: "repository after the terminator", args: []string{"--", "a/b"}, wantRepo: "a/b"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseApply(tc.args)
			if err != nil {
				t.Fatalf("parseApply(%q) error: %v", tc.args, err)
			}
			if got.repo.String() != tc.wantRepo || got.check != tc.wantCheck {
				t.Errorf("parseApply(%q) = (%s, %t), want (%s, %t)", tc.args, got.repo, got.check, tc.wantRepo, tc.wantCheck)
			}
		})
	}
}

func TestParseApplyInvalid(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"a/b", "c/d"},
		{"--", "a/b", "--check"},
		{"../.."},
		{"--force", "a/b"},
	} {
		if got, err := parseApply(args); err == nil {
			t.Errorf("parseApply(%q) = %+v, want an error", args, got)
		}
	}
}

func TestVersion(t *testing.T) {
	// Go reports the module or pseudo-version from version control, or (devel) without version control information.
	want := regexp.MustCompile(`^harness (\(devel\)|v\d+\.\d+\.\d+\S*)\n$`)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--version"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("run(--version) = %d with stderr %q, want 0 and no stderr", code, stderr.String())
	}
	if !want.MatchString(stdout.String()) {
		t.Errorf("run(--version) stdout = %q, want it to match %s", stdout.String(), want)
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
