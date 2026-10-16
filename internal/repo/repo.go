// Names GitHub repositories as owner/repo.
package repo

import (
	"fmt"
	"regexp"
	"strings"
)

// The longest account name GitHub allows.
const maxOwnerLength = 39

var (
	ownerPattern = regexp.MustCompile(`(?i)^[a-z\d]+(?:-[a-z\d]+)*$`)
	repoPattern  = regexp.MustCompile(`^[-.\w]+$`)
)

// Identifies a repository by its owner and repository name.
type Name struct {
	Owner, Repo string
}

// Writes the name as owner/repo.
func (n Name) String() string {
	return n.Owner + "/" + n.Repo
}

// Reads owner/repo under GitHub's account and repository name rules.
func Parse(s string) (Name, error) {
	owner, name, found := strings.Cut(s, "/")
	if !found || !ValidOwner(owner) || !ValidRepo(name) {
		return Name{}, fmt.Errorf("repository %q is not a valid owner/repo", s)
	}
	return Name{Owner: owner, Repo: name}, nil
}

// Accepts letters, digits and single inner hyphens, up to 39 characters.
func ValidOwner(owner string) bool {
	return len(owner) <= maxOwnerLength && ownerPattern.MatchString(owner)
}

// Accepts letters, digits, '.', '_' and '-', other than the names . and ..
func ValidRepo(name string) bool {
	return repoPattern.MatchString(name) && name != "." && name != ".."
}
