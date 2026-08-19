package doguv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaultyParseVersionComparators(t *testing.T) {
	_, err := ParseVersionComparator("<=>1.2.3-1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot contain more than two characters")

	_, err = ParseVersionComparator("!^1.2.3-1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse version")
}

// check that the VersionComparator is created correctly
func createVersionComparator(t *testing.T, version string) VersionComparator {
	vC, err := ParseVersionComparator(version)
	assert.NotNil(t, vC)
	assert.Nil(t, err)
	return vC
}

// check that the Allows method is called correctly
func allows(t *testing.T, vC VersionComparator, version string) bool {
	comparisonResult, err := vC.Allows(parse(t, version))
	assert.NotNil(t, comparisonResult)
	assert.Nil(t, err)
	return comparisonResult
}

func TestLessOperator(t *testing.T) {
	v := "<1.2.3-4"
	vC := createVersionComparator(t, v)

	assert.False(t, allows(t, vC, "1.2.3-4"))
	assert.False(t, allows(t, vC, "2.2.3-4"))
	assert.True(t, allows(t, vC, "0.2.3-4"))
}

func TestGreaterOperator(t *testing.T) {
	v := ">1.2.3-4"
	vC := createVersionComparator(t, v)

	assert.False(t, allows(t, vC, "1.2.3-4"))
	assert.True(t, allows(t, vC, "2.2.3-4"))
	assert.False(t, allows(t, vC, "0.2.3-4"))
}

func TestLessOrEqualOperator(t *testing.T) {
	v := "<=1.2.3-4"
	vC := createVersionComparator(t, v)

	assert.True(t, allows(t, vC, "1.2.3-4"))
	assert.False(t, allows(t, vC, "2.2.3-4"))
	assert.True(t, allows(t, vC, "0.2.3-4"))
}

func TestGreaterOrEqualOperator(t *testing.T) {
	v := ">=1.2.3-4"
	vC := createVersionComparator(t, v)

	assert.True(t, allows(t, vC, "1.2.3-4"))
	assert.True(t, allows(t, vC, "2.2.3-4"))
	assert.False(t, allows(t, vC, "0.2.3-4"))
}

func TestEqualOperator(t *testing.T) {
	v := "=1.2.3-4"
	vC := createVersionComparator(t, v)

	assert.True(t, allows(t, vC, "1.2.3-4"))
	assert.False(t, allows(t, vC, "2.2.3-4"))
	assert.False(t, allows(t, vC, "0.2.3-4"))
}

func TestWithEmptyVersion(t *testing.T) {
	v := ""
	_, err := ParseVersionComparator(v)
	assert.Nil(t, err)
}
