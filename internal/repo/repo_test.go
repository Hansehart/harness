package repo

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Name
	}{
		{in: "Hansehart/harness", want: Name{Owner: "Hansehart", Repo: "harness"}},
		{in: "a/b", want: Name{Owner: "a", Repo: "b"}},
		{in: "octo-org/repo.name_1", want: Name{Owner: "octo-org", Repo: "repo.name_1"}},
		{in: strings.Repeat("a", 39) + "/x", want: Name{Owner: strings.Repeat("a", 39), Repo: "x"}},
	} {
		got, err := Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
		if got.String() != tc.in {
			t.Errorf("Parse(%q).String() = %q, want %q", tc.in, got.String(), tc.in)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{
		"",
		"harness",
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
		"owner/repo/extra",
	} {
		if got, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) = %+v, want an error", in, got)
		} else if !strings.Contains(err.Error(), in) {
			t.Errorf("Parse(%q) error %q, want it to name the input", in, err)
		}
	}
}

func TestValidOwnerAndRepo(t *testing.T) {
	if !ValidOwner("octo-org") || ValidOwner("octo--org") {
		t.Error("ValidOwner accepts single hyphens only")
	}
	if !ValidRepo("repo.name_1") || ValidRepo("..") || ValidRepo("") {
		t.Error("ValidRepo accepts GitHub repository names other than . and ..")
	}
}
