package doguv3

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// versionParser holds the temporary state during parsing to avoid deep nesting and repetitive error handling.
type versionParser struct {
	raw string
	err error
}

func ParseVersion(raw string) (Version, error) {
	v := Version{Raw: raw}
	p := &versionParser{raw: removeOperatorFromRawIfPresent(raw)}

	// 1. Split into semver and extra components
	mainParts := strings.Split(p.raw, "-")
	if len(mainParts) > 2 {
		return v, fmt.Errorf("found more than one hyphen in version %s", p.raw)
	}

	// 2. Parse the primary semver components (Major.Minor.Patch.Nano)
	semverPlusNano := strings.Split(mainParts[0], ".")
	v.Major = p.parseInt(semverPlusNano, 0, "major")
	v.Minor = p.parseInt(semverPlusNano, 1, "minor")
	v.Patch = p.parseInt(semverPlusNano, 2, "patch")
	v.Nano = p.parseInt(semverPlusNano, 3, "nano")

	// 3. Parse the extra component if it exists
	if len(mainParts) > 1 {
		v.Extra = p.parseInt(mainParts, 1, "extra")
	}

	// 4. Return the first error encountered, if any
	if p.err != nil {
		return Version{Raw: raw}, p.err
	}

	return v, nil
}

// parseInt safely extracts a numeric component by index and tracks parsing errors.
func (p *versionParser) parseInt(parts []string, index int, name string) int {
	// If a previous step failed or the index doesn't exist, skip and do nothing
	if p.err != nil || index >= len(parts) {
		return 0
	}

	val, err := strconv.Atoi(parts[index])
	if err != nil {
		p.err = fmt.Errorf("failed to parse %s version %s: %w", name, parts[index], err)
		return 0
	}

	return val
}

// ParseVersion parses a raw version string and returns a Version struct

func removeOperatorFromRawIfPresent(raw string) string {
	r, _ := regexp.Compile(operatorRegex)
	if r.MatchString(raw) {
		idx := r.FindStringIndex(raw)
		raw = raw[idx[1]:] //remove operator from raw
	}
	return raw
}

// Version struct can be used to extract single parts of a version number or to compare version with each other.
// The version struct can with four or fewer digits, plus an extra version which is divided by a hyphen.
// For example: 4.0.7.11-3 => 4 Major, 0 Minor, 7 Patch, 11 Nano, 3 Extra
type Version struct {
	Raw   string
	Major int
	Minor int
	Patch int
	Nano  int
	Extra int
}

type comparisonResult int

const (
	older comparisonResult = -1
	equal comparisonResult = 0
	newer comparisonResult = 1
)

// IsNewerThan returns true if this version is newer than the given version parameter
func (v *Version) IsNewerThan(o Version) bool {
	return v.compare(o) == newer
}

// IsOlderThan returns true if this version is older than the given version parameter
func (v *Version) IsOlderThan(o Version) bool {
	return v.compare(o) == older
}

// IsEqualTo returns true if this version is equal to the given version parameter, otherwise false.
func (v *Version) IsEqualTo(o Version) bool {
	return v.compare(o) == equal
}

// IsOlderOrEqualThan returns true if this version is older or equal than the given version parameter
func (v *Version) IsOlderOrEqualThan(o Version) bool {
	result := v.compare(o)
	return result == older || result == equal
}

// IsNewerOrEqualThan returns true if this version is newer than the given version parameter
func (v *Version) IsNewerOrEqualThan(o Version) bool {
	result := v.compare(o)
	return result == newer || result == equal
}

func (v *Version) compare(o Version) comparisonResult {
	parts := v.getParts()
	otherParts := o.getParts()
	for i := 0; i < 5; i++ {
		if parts[i] > otherParts[i] {
			return comparisonResult(newer)
		} else if parts[i] < otherParts[i] {
			return comparisonResult(older)
		}
	}
	return comparisonResult(equal)
}

func (v *Version) getParts() []int {
	return []int{v.Major, v.Minor, v.Patch, v.Nano, v.Extra}
}

// String returns a string representation of a Version object. The string will be reduced to the format Major.Minor.Patch
// if the values of Extra and/or Nano are equal to zero and thus indicating that they are not set.
func (v *Version) String() string {
	if v.Raw != "" {
		return v.Raw
	}

	verBuilder := strings.Builder{}
	verBuilder.WriteString(strconv.Itoa(v.Major))
	verBuilder.WriteString(".")
	verBuilder.WriteString(strconv.Itoa(v.Minor))
	verBuilder.WriteString(".")
	verBuilder.WriteString(strconv.Itoa(v.Patch))
	if v.Nano != 0 || v.Extra != 0 {
		verBuilder.WriteString(".")
		verBuilder.WriteString(strconv.Itoa(v.Nano))
	}
	if v.Extra != 0 {
		verBuilder.WriteString("-")
		verBuilder.WriteString(strconv.Itoa(v.Extra))
	}

	return verBuilder.String()
}

// ByVersion implements sort.Interface for []Version to sort versions
type ByVersion []Version

// Len is the number of elements in the collection.
func (versions ByVersion) Len() int {
	return len(versions)
}

// Swap swaps the elements with indexes i and j.
func (versions ByVersion) Swap(i, j int) {
	versions[i], versions[j] = versions[j], versions[i]
}

// Less reports whether the element with index i should sort before the element with index j.
func (versions ByVersion) Less(i, j int) bool {
	isNewer := versions[i].IsNewerThan(versions[j])
	return isNewer
}
