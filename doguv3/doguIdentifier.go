package doguv3

import (
	"fmt"
	"regexp"

	"golang.org/x/mod/semver"
)

// Identifier identifies a Dogu via namespace, name and version
type Identifier struct {
	DoguNamespace string
	Name          string
	Version       string
}

// String returns the string presentation of this object. Like <doguNamespace>/<chartName>:<chartVersion>
func (n Identifier) String() string {
	return fmt.Sprintf("%s/%s:%s", n.DoguNamespace, n.Name, n.Version)
}

// IsValid checks, whether the identifier is valid. A valid identifier must have
//   - a non-empty DoguNamespace consisting of only lower case letters (a-z), numbers (0-9), minus (-) and underscores (_)
//   - a non-empty Name consisting of only lower case letters (a-z), numbers (0-9), minus (-) and underscores (_)
//   - a non-empty Version that matches the SemVer format (MAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]])
func (n Identifier) IsValid() bool {
	return IsValidName(n.DoguNamespace) && IsValidName(n.Name) && IsValidVersion(n.Version)
}

const nameRegex = "^[a-z0-9_\\-]+$"

func IsValidName(name string) bool {
	matched, _ := regexp.MatchString(nameRegex, name)
	return matched
}

func IsValidVersion(version string) bool {
	return semver.IsValid(fmt.Sprintf("v%s", version))
}
