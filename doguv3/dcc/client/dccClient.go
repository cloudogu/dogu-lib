package client

import (
	"context"

	"github.com/cloudogu/dogu-lib/doguv3"
)

type DccClient interface {
	// GetLatest returns the latest dogu descriptor for a dogu from the remote server.
	// Generic error if there is any issue
	GetLatest(ctx context.Context, doguNamespace string, name string) (*doguv3.Dogu, error)

	// Get returns a version specific dogu descriptor.
	// Generic error if there is any issue
	Get(ctx context.Context, doguIdentifier doguv3.Identifier) (*doguv3.Dogu, error)

	// GetVersions returns a version specific dogu descriptor.
	// Generic error if there is any issue
	GetVersions(ctx context.Context, doguNamespace string, name string) ([]string, error)

	// GetAll returns latest doguv3 identifiers of all dogus in the remote server.
	// Generic error if there is any issue
	GetAll(ctx context.Context) ([]doguv3.Identifier, error)
}
