// Package repo names GitHub repositories as owner/repo.
package repo

import (
	"fmt"
	"regexp"
	"strings"
)

// maxOwnerLength is the longest account name GitHub allows.
const maxOwnerLength = 39

var (
	ownerPattern = regexp.MustCompile(`(?i)^[a-z\d]+(?:-[a-z\d]+)*$`)
	repoPattern  = regexp.MustCompile(`^[-.\w]+$`)
)

// Name identifies a repository by its owner and repository name.
type Name struct {
	Owner, Repo string
}

// String writes the name as owner/repo.
func (n Name) String() string {
	return n.Owner + "/" + n.Repo
}

// Parse reads owner/repo under GitHub's account and repository name rules.
func Parse(s string) (Name, error) {
	owner, name, found := strings.Cut(s, "/")
	if !found || !ValidOwner(owner) || !ValidRepo(name) {
		return Name{}, fmt.Errorf("repository %q is not a valid owner/repo", s)
	}
	return Name{Owner: owner, Repo: name}, nil
}

// ValidOwner accepts letters, digits and single inner hyphens, up to 39 characters.
func ValidOwner(owner string) bool {
	return len(owner) <= maxOwnerLength && ownerPattern.MatchString(owner)
}

// ValidRepo accepts letters, digits, '.', '_' and '-', other than the names . and ..
func ValidRepo(name string) bool {
	return repoPattern.MatchString(name) && name != "." && name != ".."
}
