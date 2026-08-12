package doguv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid_Valid(t *testing.T) {
	validIdentifiers := []Identifier{
		{
			DoguNamespace: "official",
			ChartName:     "dogu",
			ChartVersion:  "1.0.0",
		},
	}
	for _, identifier := range validIdentifiers {
		assert.True(t, identifier.IsValid())
	}
}
func TestIsValid_Invalid(t *testing.T) {
	tests := []struct {
		name       string
		identifier *Identifier
	}{
		{
			name:       "nil identifier",
			identifier: nil,
		},
		{
			name:       "empty identifier",
			identifier: &Identifier{},
		},
		{
			name: "missing namespace",
			identifier: &Identifier{
				DoguNamespace: "",
				ChartName:     "dogu",
				ChartVersion:  "1.0.0",
			},
		},
		{
			name: "namespace with slash",
			identifier: &Identifier{
				DoguNamespace: "name/space",
				ChartName:     "dogu",
				ChartVersion:  "1.0.0",
			},
		},
		{
			name: "missing dogu name",
			identifier: &Identifier{
				DoguNamespace: "namespace",
				ChartName:     "",
				ChartVersion:  "1.0.0",
			},
		},
		{
			name: "dogu name with slash",
			identifier: &Identifier{
				DoguNamespace: "namespace",
				ChartName:     "dogu/name",
				ChartVersion:  "1.0.0",
			},
		},
		{
			name: "missing version",
			identifier: &Identifier{
				DoguNamespace: "namespace",
				ChartName:     "dogu",
				ChartVersion:  "",
			},
		},
		{
			name: "invalid version",
			identifier: &Identifier{
				DoguNamespace: "namespace",
				ChartName:     "dogu",
				ChartVersion:  "1.2.3.4",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.False(t, test.identifier.IsValid())
		})
	}
}

func TestString(t *testing.T) {
	identifier := Identifier{
		DoguNamespace: "namespace",
		ChartName:     "dogu",
		ChartVersion:  "1.0.2",
	}
	assert.Equal(t, "namespace/dogu:1.0.2", identifier.String())
}

func TestString_empty(t *testing.T) {
	var identifier Identifier
	assert.Equal(t, "/:", identifier.String())
}

func TestString_nil(t *testing.T) {
	var identifier *Identifier
	assert.Equal(t, "nil", identifier.String())
}
