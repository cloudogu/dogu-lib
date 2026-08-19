package doguv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptyVersion(t *testing.T) {
	_, err := ParseVersion("")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse major version ")
}

func TestParseVersion(t *testing.T) {
	version, err := ParseVersion("4.2.7.11-3")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 2, version.Minor)
	assert.Equal(t, 7, version.Patch)
	assert.Equal(t, 11, version.Nano)
	assert.Equal(t, 3, version.Extra)
}

func TestParseVersionWithoutNano(t *testing.T) {
	version, err := ParseVersion("4.2.7-3")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 2, version.Minor)
	assert.Equal(t, 7, version.Patch)
	assert.Equal(t, 0, version.Nano)
	assert.Equal(t, 3, version.Extra)
}

func TestParseVersionWithoutPatch(t *testing.T) {
	version, err := ParseVersion("4.2-3")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 2, version.Minor)
	assert.Equal(t, 0, version.Patch)
	assert.Equal(t, 0, version.Nano)
	assert.Equal(t, 3, version.Extra)
}

func TestParseVersionWithoutMinor(t *testing.T) {
	version, err := ParseVersion("4-3")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 0, version.Minor)
	assert.Equal(t, 0, version.Patch)
	assert.Equal(t, 0, version.Nano)
	assert.Equal(t, 3, version.Extra)
}

func TestParseVersionWithoutExtra(t *testing.T) {
	version, err := ParseVersion("4.2")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 2, version.Minor)
	assert.Equal(t, 0, version.Patch)
	assert.Equal(t, 0, version.Nano)
	assert.Equal(t, 0, version.Extra)
}

func TestParseVersionOnlyMajor(t *testing.T) {
	version, err := ParseVersion("4")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 0, version.Minor)
	assert.Equal(t, 0, version.Patch)
	assert.Equal(t, 0, version.Nano)
	assert.Equal(t, 0, version.Extra)
}

func TestParseVersionAndIgnoreUnknownParts(t *testing.T) {
	version, err := ParseVersion("4.2.7.11.23-3")
	assert.Nil(t, err)
	assert.Equal(t, 4, version.Major)
	assert.Equal(t, 2, version.Minor)
	assert.Equal(t, 7, version.Patch)
	assert.Equal(t, 11, version.Nano)
	assert.Equal(t, 3, version.Extra)
}

func TestParseVersionWithNonNumber(t *testing.T) {
	_, err := ParseVersion("4.trillian.3")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "trillian")
	assert.Contains(t, err.Error(), "minor")
}

func TestIsNewerThan(t *testing.T) {
	assertIsNewerThan(t, "5", "4")
	assertIsNewerThan(t, "5", "4.11")
	assertIsNewerThan(t, "4.12", "4.11")
	assertIsNewerThan(t, "4.12.3.1", "4.12.3")
	assertIsNewerThan(t, "4.12.3.1-2", "4.12.3.1")

	assertNegativeIsNewerThan(t, "4.9.12.1-2", "4.10.0.1-2")

	assertNegativeIsNewerThan(t, "4", "5")
	assertNegativeIsNewerThan(t, "4.11", "5")
	assertNegativeIsNewerThan(t, "4.11", "4.12")
	assertNegativeIsNewerThan(t, "4.12.3", "4.12.3.1")
	assertNegativeIsNewerThan(t, "4.12.3.1", "4.12.3.1-2")
	assertNegativeIsNewerThan(t, "4.12.3.1-2", "4.12.3.1-2")

}

func TestIsOlderOrEqualThan(t *testing.T) {
	assertIsOlderOrEqualThan(t, "4", "5")
	assertIsOlderOrEqualThan(t, "4.11", "5")
	assertIsOlderOrEqualThan(t, "4.11", "4.12")
	assertIsOlderOrEqualThan(t, "4.12.3", "4.12.3.1")
	assertIsOlderOrEqualThan(t, "4.12.3.1", "4.12.3.1-2")
	assertIsOlderOrEqualThan(t, "4.12.3.1-2", "4.12.3.1-2")

	assertNegativeIsOlderOrEqualThan(t, "4.14.1.2-8", "4.12.1.10-16")

	assertNegativeIsOlderOrEqualThan(t, "5", "4")
	assertNegativeIsOlderOrEqualThan(t, "5", "4.11")
	assertNegativeIsOlderOrEqualThan(t, "4.12", "4.11")
	assertNegativeIsOlderOrEqualThan(t, "4.12.3.1", "4.12.3")
	assertNegativeIsOlderOrEqualThan(t, "4.12.3.1-2", "4.12.3.1")
}

func TestParseVersionWithOperatorGreaterThan(t *testing.T) {
	version, err := ParseVersion(">1.17.10-7")
	assert.NotNil(t, version)
	assert.Nil(t, err)
	assert.Equal(t, version.Major, 1)
	assert.Equal(t, version.Minor, 17)
	assert.Equal(t, version.Patch, 10)
	assert.Equal(t, version.Extra, 7)
}

func TestParseVersionWithOperatorLessThan(t *testing.T) {
	version, err := ParseVersion("<1.17.10-7")
	assert.NotNil(t, version)
	assert.Nil(t, err)
	assert.Equal(t, version.Major, 1)
	assert.Equal(t, version.Minor, 17)
	assert.Equal(t, version.Patch, 10)
	assert.Equal(t, version.Extra, 7)
}

func TestParseVersionWithOperatorGreaterEqualThan(t *testing.T) {
	version, err := ParseVersion(">=1.17.10-7")
	assert.NotNil(t, version)
	assert.Nil(t, err)
	assert.Equal(t, version.Major, 1)
	assert.Equal(t, version.Minor, 17)
	assert.Equal(t, version.Patch, 10)
	assert.Equal(t, version.Extra, 7)
}

func TestParseVersionWithOperatorLessEqualThan(t *testing.T) {
	version, err := ParseVersion("<=1.17.10-7")
	assert.NotNil(t, version)
	assert.Nil(t, err)
	assert.Equal(t, version.Major, 1)
	assert.Equal(t, version.Minor, 17)
	assert.Equal(t, version.Patch, 10)
	assert.Equal(t, version.Extra, 7)
}

func assertNegativeIsNewerThan(t *testing.T, s1, s2 string) {
	version := parse(t, s1)
	otherVersion := parse(t, s2)
	assert.False(t, version.IsNewerThan(otherVersion), "version %s should not be newer than %s", s1, s2)
	assert.False(t, otherVersion.IsOlderThan(version), "version %s should not be newer than %s", s1, s2)
}

func assertNegativeIsOlderOrEqualThan(t *testing.T, s1, s2 string) {

	version1, _ := ParseVersion(s1)
	version2, _ := ParseVersion(s2)
	assert.False(t, version1.IsOlderOrEqualThan(version2), "version %s should not be older or equal than %s", s1, s2)
}

func TestEquals(t *testing.T) {
	left, err := ParseVersion("4.12.3.1-2")
	assert.Nil(t, err)
	right, err := ParseVersion("4.12.3.1-2")
	assert.Nil(t, err)

	assert.False(t, left.IsNewerThan(right))
	assert.False(t, right.IsNewerThan(left))
	assert.False(t, left.IsOlderThan(right))
	assert.False(t, right.IsOlderThan(left))

	assert.True(t, left.IsOlderOrEqualThan(right))
	assert.True(t, right.IsOlderOrEqualThan(left))
}

func assertIsNewerThan(t *testing.T, rawVersion string, otherRawVersion string) {
	version := parse(t, rawVersion)
	otherVersion := parse(t, otherRawVersion)
	assert.True(t, version.IsNewerThan(otherVersion), "version %s should be newer than %s", rawVersion, otherRawVersion)
	assert.True(t, otherVersion.IsOlderThan(version), "version %s should be newer than %s", rawVersion, otherRawVersion)
}

func parse(t *testing.T, raw string) Version {
	v, err := ParseVersion(raw)
	assert.Nil(t, err)
	return v
}

func assertIsOlderOrEqualThan(t *testing.T, rawVersion string, otherRawVersion string) {
	version := parse(t, rawVersion)
	otherVersion := parse(t, otherRawVersion)
	assert.True(t, version.IsOlderOrEqualThan(otherVersion), "version %s should be older or equal than %s", rawVersion, otherRawVersion)
}

func TestLess(t *testing.T) {
	assertLess(t, "5", "4")
	assertLess(t, "5", "4.11")
	assertLess(t, "4.12", "4.11")
	assertLess(t, "4.12.3.1", "4.12.3")
	assertLess(t, "4.12.3.1-2", "4.12.3.1")

	assertNegativeLess(t, "4", "5")
	assertNegativeLess(t, "4.11", "5")
	assertNegativeLess(t, "4.11", "4.12")
	assertNegativeLess(t, "4.12.3", "4.12.3.1")
	assertNegativeLess(t, "4.12.3.1", "4.12.3.1-2")
	assertNegativeLess(t, "4.12.3.1-2", "4.12.3.1-2")
}

func assertLess(t *testing.T, s1, s2 string) {

	v1 := parse(t, s1)
	v2 := parse(t, s2)

	var byVersion ByVersion
	byVersion = append(byVersion, v1)
	byVersion = append(byVersion, v2)

	assert.True(t, byVersion.Less(0, 1), "version %s should be newer than %s", v1.Raw, v2.Raw)
}

func assertNegativeLess(t *testing.T, s1, s2 string) {

	v1 := parse(t, s1)
	v2 := parse(t, s2)

	var byVersion ByVersion
	byVersion = append(byVersion, v1)
	byVersion = append(byVersion, v2)

	assert.False(t, byVersion.Less(0, 1), "version %s should be older than %s", v1.Raw, v2.Raw)
}

func TestIsNewerOrEqualThan(t *testing.T) {
	// newer
	assertIsNewerOrEqualThan(t, "5", "4")
	assertIsNewerOrEqualThan(t, "5", "4.11")
	assertIsNewerOrEqualThan(t, "4.12", "4.11")
	assertIsNewerOrEqualThan(t, "4.12.3.1", "4.12.3")
	assertIsNewerOrEqualThan(t, "4.12.3.1-2", "4.12.3.1")
	assertIsNewerOrEqualThan(t, "4.14.1.2-8", "4.12.1.10-16")

	// equal
	assertIsNewerOrEqualThan(t, "4.12.3.1-2", "4.12.3.1-2")

	assertNegativeIsNewerOrEqualThan(t, "4", "5")
	assertNegativeIsNewerOrEqualThan(t, "4.11", "5")
	assertNegativeIsNewerOrEqualThan(t, "4.11", "4.12")
	assertNegativeIsNewerOrEqualThan(t, "4.12.3", "4.12.3.1")
	assertNegativeIsNewerOrEqualThan(t, "4.12.3.1", "4.12.3.1-2")
	assertNegativeIsNewerOrEqualThan(t, "4.12.1.10-16", "4.14.1.2-8")
}

func assertIsNewerOrEqualThan(t *testing.T, rawVersion string, otherRawVersion string) {
	version := parse(t, rawVersion)
	otherVersion := parse(t, otherRawVersion)
	assert.True(t, version.IsNewerOrEqualThan(otherVersion), "version %s should be newer or equal than %s", rawVersion, otherRawVersion)
}

func assertNegativeIsNewerOrEqualThan(t *testing.T, s1, s2 string) {
	version1, _ := ParseVersion(s1)
	version2, _ := ParseVersion(s2)
	assert.False(t, version1.IsNewerOrEqualThan(version2), "version %s should not be newer or equal than %s", s1, s2)
}

func TestVersion_String(t *testing.T) {
	type fields struct {
		Raw   string
		Major int
		Minor int
		Patch int
		Nano  int
		Extra int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"return raw", fields{Raw: "1.2.3.4-5"}, "1.2.3.4-5"},
		{"return 1-part version", fields{Major: 1}, "1.0.0"},
		{"return 2-part version", fields{Major: 1, Minor: 2}, "1.2.0"},
		{"return 3-part version", fields{Major: 1, Minor: 2, Patch: 3}, "1.2.3"},
		{"return 4-part version", fields{Major: 1, Minor: 2, Patch: 3, Nano: 4}, "1.2.3.4"},
		{"return 5-part version", fields{Major: 1, Minor: 2, Patch: 3, Nano: 4, Extra: 5}, "1.2.3.4-5"},
		{"return version for empty", fields{}, "0.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Raw:   tt.fields.Raw,
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
				Nano:  tt.fields.Nano,
				Extra: tt.fields.Extra,
			}
			assert.Equalf(t, tt.want, v.String(), "String()")
		})
	}
}
