// Repository names the hosting service accepts.
package repo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// The hosting service's limit for account names.
const maxOwnerLength = 39

// Marks a name the hosting service would reject.
var ErrInvalid = errors.New("not a valid owner/repo")

var (
	ownerPattern = regexp.MustCompile(`(?i)^[a-z\d]+(?:-[a-z\d]+)*$`)
	repoPattern  = regexp.MustCompile(`^[-.\w]+$`)
)

// A repository, addressed through its owner.
type Name struct {
	Owner, Repo string
}

// The form people type and the service displays.
func (n Name) String() string {
	return n.Owner + "/" + n.Repo
}

// Accepts only names the hosting service would accept.
func Parse(s string) (Name, error) {
	owner, name, found := strings.Cut(s, "/")
	if !found || !ValidOwner(owner) || !ValidRepo(name) {
		return Name{}, fmt.Errorf("repository %q: %w", s, ErrInvalid)
	}

	return Name{Owner: owner, Repo: name}, nil
}

// Applies the service's rules for account names.
func ValidOwner(owner string) bool {
	return len(owner) <= maxOwnerLength && ownerPattern.MatchString(owner)
}

// Applies the service's rules for repository names.
func ValidRepo(name string) bool {
	return repoPattern.MatchString(name) && name != "." && name != ".."
}
