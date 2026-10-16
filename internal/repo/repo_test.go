package repo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Hansehart/harness/internal/repo"
)

// The longest account name GitHub allows.
const longestOwner = 39

func TestParse(t *testing.T) {
	t.Parallel()

	longest := strings.Repeat("a", longestOwner)
	tests := []struct {
		input string
		want  repo.Name
	}{
		{
			input: "Hansehart/harness",
			want:  repo.Name{Owner: "Hansehart", Repo: "harness"},
		},
		{input: "a/b", want: repo.Name{Owner: "a", Repo: "b"}},
		{
			input: "octo-org/repo.name_1",
			want:  repo.Name{Owner: "octo-org", Repo: "repo.name_1"},
		},
		{input: longest + "/x", want: repo.Name{Owner: longest, Repo: "x"}},
	}

	for _, test := range tests {
		got, err := repo.Parse(test.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", test.input, err)

			continue
		}

		if got != test.want || got.String() != test.input {
			t.Errorf("Parse(%q) = %+v, want %+v", test.input, got, test.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()

	inputs := []string{
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
		strings.Repeat("o", longestOwner+1) + "/repo",
		"a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a-a/repo",
		"owner/re?po",
		"owner/repo/extra",
	}

	for _, input := range inputs {
		got, err := repo.Parse(input)
		if !errors.Is(err, repo.ErrInvalid) {
			t.Errorf("Parse(%q) = %+v, %v; want ErrInvalid", input, got, err)

			continue
		}

		if !strings.Contains(err.Error(), input) {
			t.Errorf("Parse(%q) error %q, want the input in it", input, err)
		}
	}
}

func TestValidOwnerAndRepo(t *testing.T) {
	t.Parallel()

	if !repo.ValidOwner("octo-org") || repo.ValidOwner("octo--org") {
		t.Error("ValidOwner accepts single hyphens only")
	}

	if !repo.ValidRepo("repo.name_1") {
		t.Error("ValidRepo rejects letters, digits and '.', '_', '-'")
	}

	if repo.ValidRepo("..") || repo.ValidRepo("") {
		t.Error("ValidRepo accepts the path name .. or an empty name")
	}
}
