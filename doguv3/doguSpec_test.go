package doguv3_test

import (
	"testing"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDoguClone(t *testing.T) {
	originalDogu := testutil.CreateTestDoguV3()

	clone := originalDogu.Clone()

	assert.Equal(t, originalDogu, clone)
}

func TestDoguCloneNil(t *testing.T) {
	var originalDogu *doguv3.Dogu

	clone := originalDogu.Clone()

	assert.Equal(t, originalDogu, clone)
}
