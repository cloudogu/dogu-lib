package doguv3

import (
	"regexp"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// go does not support lookaheads or lookbehinds (tried "^[=,>,<]+[\"d\"-,.]*" "^[=,>,<]+(?=[d])+"
const operatorRegex = `^[=><]+`

// =, ==, <, >, <=, >=
const (
	// version compare operators
	operatorEqual              = "="
	operatorEqualDouble        = "=="
	operatorLessThan           = "<"
	operatorGreaterThan        = ">"
	operatorLessOrEqualThan    = "<="
	operatorGreaterOrEqualThan = ">="
)

type operator string

// VersionComparator is responsible to compare versions and to check defined constraints.
type VersionComparator struct {
	version  Version
	operator operator
}

// ParseVersionComparator creates a new version comparator by parsing a raw version string.
func ParseVersionComparator(raw string) (VersionComparator, error) {
	var version Version

	operator, err := parseOperator(raw)
	if err != nil {
		return VersionComparator{}, errors.Wrap(err, "failed to parse operator")
	}

	if raw == "" {
		return VersionComparator{}, nil
	}

	version, err = ParseVersion(raw[len(operator):])
	if err != nil {
		return VersionComparator{}, errors.Wrap(err, "failed to parse version")
	}
	return VersionComparator{
		version:  version,
		operator: operator,
	}, nil
}

// Allows check whether the given version fulfills the requirements for the version comparator.
func (v VersionComparator) Allows(version Version) (bool, error) {
	switch v.operator {
	case operatorEqual, operatorEqualDouble: //Approximately equivalent to version
		return v.version.IsEqualTo(version), nil
	case operatorGreaterThan:
		return v.version.IsOlderThan(version), nil
	case operatorLessThan:
		return v.version.IsNewerThan(version), nil
	case operatorGreaterOrEqualThan:
		return v.version.IsOlderOrEqualThan(version), nil
	case operatorLessOrEqualThan:
		return v.version.IsNewerOrEqualThan(version), nil
	case "":
		// this is the edge-case that the dogu.json has no version field or an empty version. At this point we allow
		// every version. if a list of dogus is used it should be best practice to sort the list by version
		if v.version.Raw == "" {
			return true, nil
		}
		return v.version.IsEqualTo(version), nil
	default:
		/* We could use the default operator here but this is a conscious choice to not do this,
		as there are places in the calling code where the error is caught and assessed accordingly */
		return false, errors.Errorf("could not find suitable comperator for '%s' operator", v.operator)
	}
}

func parseOperator(raw string) (operator, error) {
	var op operator

	r, _ := regexp.Compile(operatorRegex)
	if r.MatchString(raw) {
		idx := r.FindStringIndex(raw)
		op = operator(raw[idx[0]:idx[1]]) //cut operator
		if len(op) > 2 {
			err := errors.Errorf("dependency operator %s of Version %s cannot contain more than two characters. Allowed operators are =,==,>,<,>= and <=", op, raw)
			logrus.Error(err)
			return "", err
		}
	}
	if len(op) == 0 {
		logrus.Debug(errors.Errorf("no dependency operator in %s could be found", raw))
		return "", nil
	}
	return op, nil
}
