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
		{name: "apply without repository", args: []string{"apply"}, wantCode: 2, wantStderr: "usage:"},
		{name: "apply with two repositories", args: []string{"apply", "a/b", "c/d"}, wantCode: 2, wantStderr: "usage:"},
		{name: "apply with a malformed repository", args: []string{"apply", "harness"}, wantCode: 2, wantStderr: "owner/repo"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tc.args, &stdout, &stderr)
			if got != tc.wantCode {
				t.Errorf("run(%q) = %d, want %d", tc.args, got, tc.wantCode)
			}
			if tc.wantStdout != "" && stdout.String() != tc.wantStdout {
				t.Errorf("run(%q) stdout = %q, want %q", tc.args, stdout.String(), tc.wantStdout)
			}
			if !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("run(%q) stderr = %q, want it to contain %q", tc.args, stderr.String(), tc.wantStderr)
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
