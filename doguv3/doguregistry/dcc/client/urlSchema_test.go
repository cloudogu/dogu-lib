package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/config"
	"github.com/cloudogu/dogu-lib/doguv3/internal/testutil"
)

func Test_newURLSchema(t *testing.T) {
	baseURL, err := url.Parse("https://dogu.cloudogu.com/api/v3/dogus")
	require.NoError(t, err)

	identifier := doguv3.Identifier{DoguNamespace: "official", Name: "redmine", Version: "0.0.1"}

	tests := []struct {
		name            string
		schemaName      string
		wantGetAll      string
		wantGetLatest   string
		wantGet         string
		wantGetVersions string
	}{
		{
			name:            "default schema addresses the API routes",
			schemaName:      config.URLSchemaDefault,
			wantGetAll:      "https://dogu.cloudogu.com/api/v3/dogus",
			wantGetLatest:   "https://dogu.cloudogu.com/api/v3/dogus/official/redmine",
			wantGet:         "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/0.0.1",
			wantGetVersions: "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/_versions",
		},
		{
			name:            "empty schema falls back to default",
			schemaName:      "",
			wantGetAll:      "https://dogu.cloudogu.com/api/v3/dogus",
			wantGetLatest:   "https://dogu.cloudogu.com/api/v3/dogus/official/redmine",
			wantGet:         "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/0.0.1",
			wantGetVersions: "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/_versions",
		},
		{
			name:            "index schema addresses files",
			schemaName:      config.URLSchemaIndex,
			wantGetAll:      "https://dogu.cloudogu.com/api/v3/dogus/index.json",
			wantGetLatest:   "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/index.json",
			wantGet:         "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/0.0.1/index.json",
			wantGetVersions: "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/_versions.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, err := newURLSchema(baseURL, tt.schemaName)

			require.NoError(t, err)
			assert.Equal(t, tt.wantGetAll, sut.getAll())
			assert.Equal(t, tt.wantGetLatest, sut.getLatest("official", "redmine"))
			assert.Equal(t, tt.wantGet, sut.get(identifier))
			assert.Equal(t, tt.wantGetVersions, sut.getVersions("official", "redmine"))
		})
	}

	t.Run("should return an error for an unknown schema", func(t *testing.T) {
		sut, err := newURLSchema(baseURL, "webdav")

		require.Error(t, err)
		assert.Nil(t, sut)
		assert.ErrorContains(t, err, "unknown url schema 'webdav', expected 'default' or 'index'")
		assert.True(t, doguregistry.IsGenericError(err))
	})

	t.Run("should keep a base URL without a path", func(t *testing.T) {
		rootURL, err := url.Parse("https://dogu.cloudogu.com")
		require.NoError(t, err)

		sut, err := newURLSchema(rootURL, config.URLSchemaIndex)

		require.NoError(t, err)
		assert.Equal(t, "https://dogu.cloudogu.com/index.json", sut.getAll())
		assert.Equal(t, "https://dogu.cloudogu.com/official/redmine/index.json", sut.getLatest("official", "redmine"))
	})
}

func TestNew_urlSchema(t *testing.T) {
	t.Run("should fail for an unknown url schema", func(t *testing.T) {
		got, err := New(&config.DoguRegistryConfiguration{
			BaseURL:   "https://dogu.cloudogu.com/api/v3/dogus",
			URLSchema: "nope",
		}, nil)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "unknown url schema 'nope'")
	})
}

// TestDccHttpClient_indexSchema proves that a file-based DCC can be read: every route addresses a file, so a plain
// file server can answer it. Without the index schema, the routes for all dogus and for a single dogu would address
// a directory.
func TestDccHttpClient_indexSchema(t *testing.T) {
	dogu := testutil.CreateTestDoguV3()
	identifier := doguv3.Identifier{DoguNamespace: dogu.DoguNamespace, Name: dogu.Name, Version: dogu.Version}

	// the file tree of a mirrored DCC
	files := map[string]any{
		"/index.json":                        []doguv3.Identifier{identifier},
		"/official/redmine/index.json":       dogu,
		"/official/redmine/_versions.json":   []string{dogu.Version},
		"/official/redmine/0.0.1/index.json": dogu,
	}

	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)

		content, found := files[r.URL.Path]
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(content))
	}))
	defer server.Close()

	sut, err := New(&config.DoguRegistryConfiguration{
		BaseURL:      server.URL,
		URLSchema:    config.URLSchemaIndex,
		DisableCache: true,
	}, nil)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("GetAll reads the root index file", func(t *testing.T) {
		actual, err := sut.GetAll(ctx)

		require.NoError(t, err)
		assert.Equal(t, []doguv3.Identifier{identifier}, actual)
	})

	t.Run("GetLatest reads the index file of the dogu", func(t *testing.T) {
		actual, err := sut.GetLatest(ctx, dogu.DoguNamespace, dogu.Name)

		require.NoError(t, err)
		assert.Equal(t, dogu, actual)
	})

	t.Run("GetVersions reads the versions file of the dogu", func(t *testing.T) {
		actual, err := sut.GetVersions(ctx, dogu.DoguNamespace, dogu.Name)

		require.NoError(t, err)
		assert.Equal(t, []string{dogu.Version}, actual)
	})

	t.Run("Get reads the index file of the dogu version", func(t *testing.T) {
		actual, err := sut.Get(ctx, identifier)

		require.NoError(t, err)
		assert.Equal(t, dogu, actual)
	})

	t.Run("every request addressed a file", func(t *testing.T) {
		assert.Equal(t, []string{
			"/index.json",
			"/official/redmine/index.json",
			"/official/redmine/_versions.json",
			"/official/redmine/0.0.1/index.json",
		}, requested)
	})
}
