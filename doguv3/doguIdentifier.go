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

// String return the string presentation of this object. Like <doguNamespace>/<chartName>:<chartVersion>
func (n *Identifier) String() string {
	if n == nil {
		return "nil"
	}
	return fmt.Sprintf("%s/%s:%s", n.DoguNamespace, n.ChartName, n.ChartVersion)
}

// IsValid checks, whether the identifier is valid. A valid identifier must have
//   - a non-empty DoguNamespace without any '/' character
//   - a non-empty ChartName without any '/' character
//   - a non-empty ChartVersion that matches the SemVer format
func (n *Identifier) IsValid() bool {
	return n != nil && isValidName(n.DoguNamespace) && isValidName(n.ChartName) && isValidVersion(n.ChartVersion)
}

const nameRegex = "^[a-z0-9_\\-]+$"

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(nameRegex, name)
	return matched
}

func isValidVersion(version string) bool {
	return semver.IsValid(fmt.Sprintf("v%s", version))
}
