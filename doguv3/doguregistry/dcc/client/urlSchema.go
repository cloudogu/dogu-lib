package client

import (
	"fmt"
	"net/url"
	"path"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/config"
)

const (
	defaultVersionsPath = "_versions"
	defaultVersionsFile = "_versions.json"
	defaultIndexFile    = "index.json"
)

// urlSchema builds the request URLs for a DCC.
//
// A live DCC-v3 API serves every route as its own resource. A file-based DCC, that is a DCC mirrored onto a static
// webserver, cannot do that: the routes for all dogus and for a single dogu are a directory there, and a path cannot
// be a file and a directory at the same time. The index schema therefore addresses a file inside each of those
// directories.
type urlSchema struct {
	baseURL *url.URL
	// indexFile is addressed instead of a route that is a directory in a file-based DCC. It is empty for a live
	// DCC-v3 API, where those routes are resources of their own.
	indexFile string
	// versionsFile holds the version list of a dogu.
	versionsFile string
}

func newURLSchema(baseURL *url.URL, schemaName string) (*urlSchema, error) {
	switch schemaName {
	case "", config.URLSchemaDefault:
		return &urlSchema{baseURL: baseURL, indexFile: "", versionsFile: defaultVersionsPath}, nil
	case config.URLSchemaIndex:
		return &urlSchema{baseURL: baseURL, indexFile: defaultIndexFile, versionsFile: defaultVersionsFile}, nil
	default:
		return nil, doguregistry.NewGenericError(fmt.Errorf("unknown url schema '%s', expected '%s' or '%s'",
			schemaName, config.URLSchemaDefault, config.URLSchemaIndex))
	}
}

// getAll returns the URL of the list of all dogus.
func (s *urlSchema) getAll() string {
	return s.resolve(s.indexFile)
}

// getLatest returns the URL of the latest version of a dogu.
func (s *urlSchema) getLatest(doguNamespace string, name string) string {
	return s.resolve(doguNamespace, name, s.indexFile)
}

// get returns the URL of one specific dogu version.
func (s *urlSchema) get(identifier doguv3.Identifier) string {
	return s.resolve(identifier.DoguNamespace, identifier.Name, identifier.Version, s.indexFile)
}

// getVersions returns the URL of the version list of a dogu.
func (s *urlSchema) getVersions(doguNamespace string, name string) string {
	return s.resolve(doguNamespace, name, s.versionsFile)
}

// resolve appends the given elements to the base URL. Empty elements are dropped, which is how the default schema
// leaves out the index file.
func (s *urlSchema) resolve(elements ...string) string {
	pathElements := make([]string, 0, len(elements)+1)
	pathElements = append(pathElements, s.baseURL.Path)

	for _, element := range elements {
		if element != "" {
			pathElements = append(pathElements, element)
		}
	}

	return s.baseURL.ResolveReference(&url.URL{Path: path.Join(pathElements...)}).String()
}
