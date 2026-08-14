package doguv3

import (
	"fmt"
	"regexp"

	"golang.org/x/mod/semver"
)

// Identifier identifies a Dogu via namespace, name and version
type Identifier struct {
	DoguNamespace string
	ChartName     string
	ChartVersion  string
}

// String returns the string presentation of this object. Like <doguNamespace>/<chartName>:<chartVersion>
func (n Identifier) String() string {
	return fmt.Sprintf("%s/%s:%s", n.DoguNamespace, n.ChartName, n.ChartVersion)
}

// IsValid checks, whether the identifier is valid. A valid identifier must have
//   - a non-empty DoguNamespace consisting of only lower case letters (a-z), numbers (0-9), minus (-) and underscores (_)
//   - a non-empty ChartName consisting of only lower case letters (a-z), numbers (0-9), minus (-) and underscores (_)
//   - a non-empty ChartVersion that matches the SemVer format (MAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]])
func (n Identifier) IsValid() bool {
	return isValidName(n.DoguNamespace) && isValidName(n.ChartName) && isValidVersion(n.ChartVersion)
}

const nameRegex = "^[a-z0-9_\\-]+$"

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(nameRegex, name)
	return matched
}

func isValidVersion(version string) bool {
	return semver.IsValid(fmt.Sprintf("v%s", version))
}
