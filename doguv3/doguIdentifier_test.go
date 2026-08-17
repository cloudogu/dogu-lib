package doguv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid_Valid(t *testing.T) {
	tests := []struct {
		name       string
		identifier Identifier
	}{
		{
			name: "simple",
			identifier: Identifier{
				DoguNamespace: "official",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "namespace with valid separators",
			identifier: Identifier{
				DoguNamespace: "namespace-with_separators",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "name with valid separators",
			identifier: Identifier{
				DoguNamespace: "official",
				Name:          "dogu-with_separators",
				Version:       "1.0.0",
			},
		},
		{
			name: "pre-release version",
			identifier: Identifier{
				DoguNamespace: "official",
				Name:          "dogu",
				Version:       "1.0.0-beta.2",
			},
		},
		{
			name: "version with build info",
			identifier: Identifier{
				DoguNamespace: "official",
				Name:          "dogu",
				Version:       "1.0.0+special",
			},
		},
		{
			name: "version without patch",
			identifier: Identifier{
				DoguNamespace: "official",
				Name:          "dogu",
				Version:       "1.4",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.True(t, test.identifier.IsValid())
		})
	}
}
func TestIsValid_Invalid(t *testing.T) {
	tests := []struct {
		name       string
		identifier Identifier
	}{
		{
			name:       "empty identifier",
			identifier: Identifier{},
		},
		{
			name: "missing namespace",
			identifier: Identifier{
				DoguNamespace: "",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "namespace with slash",
			identifier: Identifier{
				DoguNamespace: "name/space",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "namespace with upper case",
			identifier: Identifier{
				DoguNamespace: "NameSpace",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "namespace with space",
			identifier: Identifier{
				DoguNamespace: "name space",
				Name:          "dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "missing dogu name",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "",
				Version:       "1.0.0",
			},
		},
		{
			name: "dogu name with slash",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "dogu/name",
				Version:       "1.0.0",
			},
		},
		{
			name: "dogu name with upper case",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "Dogu",
				Version:       "1.0.0",
			},
		},
		{
			name: "dogu name with space",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "dogu name",
				Version:       "1.0.0",
			},
		},
		{
			name: "missing version",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "dogu",
				Version:       "",
			},
		},
		{
			name: "invalid version",
			identifier: Identifier{
				DoguNamespace: "namespace",
				Name:          "dogu",
				Version:       "1.2.3.4",
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
		Name:          "dogu",
		Version:       "1.0.2",
	}
	assert.Equal(t, "namespace/dogu:1.0.2", identifier.String())
}

func TestString_empty(t *testing.T) {
	var identifier Identifier
	assert.Equal(t, "/:", identifier.String())
}
