package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteDoguDescriptorRepository(t *testing.T) {
	remoteConfig := &config.DoguRegistryConfiguration{}
	credentials := &config.Credentials{}
	got, err := New(remoteConfig, credentials)

	assert.NotNil(t, got)
	assert.Nil(t, err)
}

func TestNewRemoteDoguDescriptorRepositoryWithTimeout(t *testing.T) {
	remoteConfig := &config.DoguRegistryConfiguration{
		Timeout: 15,
	}
	credentials := &config.Credentials{}
	got, err := New(remoteConfig, credentials)

	assert.NotNil(t, got)
	assert.Nil(t, err)
}

func Test_newDccHttpClient(t *testing.T) {
	t.Run("Should return new httpRemote", func(t *testing.T) {
		remoteConfig := &config.DoguRegistryConfiguration{}
		creds := &config.Credentials{}

		_, err := New(remoteConfig, creds)

		require.NoError(t, err)
	})
	t.Run("should return error when proxy url is invalid", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{
			ProxySettings: config.ProxySettings{
				Enabled: true,
				Server:  "invalid\x7f",
				Port:    80,
			},
		}
		creds := &config.Credentials{}

		// when
		remote, err := New(remoteConfig, creds)

		// then
		require.Error(t, err)
		assert.Nil(t, remote)
		assert.Contains(t, err.Error(), "failed to parse proxy url")
	})

	t.Run("should return error when  configuration is not set ", func(t *testing.T) {
		// given

		creds := &config.Credentials{}

		// when
		remote, err := New(nil, creds)

		// then
		require.Error(t, err)
		assert.Nil(t, remote)
		assert.Contains(t, err.Error(), "dogu registry configuration must not be nil")
	})

}
func Test_New_fails_for_wrong_url(t *testing.T) {
	// given:

	remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost:8080/\n"}
	_, err := New(remoteConfig, nil)

	// then
	require.Error(t, err)
	assert.True(t, doguregistry.IsGenericError(err))
	assert.Contains(t, err.Error(), "failed to parse endpoint url")
}

func Test_createProxyHTTPTransport(t *testing.T) {
	t.Run("create transport", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{
			ProxySettings: config.ProxySettings{
				Enabled:  true,
				Server:   "1.2.3.4",
				Port:     80,
				Username: "user",
				Password: "password",
			},
		}

		// when
		transport, err := createProxyHTTPTransport(remoteConfig)
		proxy, err := transport.Proxy(nil)
		require.NoError(t, err)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://1.2.3.4:80", proxy.String())
		assert.Equal(t, "Basic dXNlcjpwYXNzd29yZA==", transport.ProxyConnectHeader.Get("Proxy-Authorization"))
	})
}

func Test_checkStatusCode(t *testing.T) {
	t.Run("should return nil for HTTP 200", func(t *testing.T) {
		mockResp := &http.Response{}
		mockResp.Status = "200 OK"
		mockResp.StatusCode = http.StatusOK
		mockResp.Body = io.NopCloser(strings.NewReader(`{"status": "is well"}`))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error for HTTP statuses >= 400", func(t *testing.T) {
		mockResp := &http.Response{}
		mockResp.Status = "405 Whoopsie!"
		mockResp.StatusCode = 405
		mockResp.Body = io.NopCloser(strings.NewReader(`{"status": "I, uh, well... phew!"}`))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.Error(t, err)
		assert.Equal(t, "remote registry returns invalid status: 405 Whoopsie!: I, uh, well... phew!: (no error)", err.Error())
	})

	t.Run("should return error for HTTP 400", func(t *testing.T) {
		const errorBody = "Do not use v1 endpoint for v2 dogu creation. Use v2 endpoint instead."

		mockResp := &http.Response{}
		mockResp.Status = http.StatusText(http.StatusBadRequest)
		mockResp.StatusCode = http.StatusBadRequest
		mockResp.Body = io.NopCloser(strings.NewReader(fmt.Sprintf(`{"error": "%s"}`, errorBody)))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.Error(t, err)
		assert.Equal(t, err.Error(), "remote registry returns invalid status: Bad Request: 400: Do not use v1 endpoint for v2 dogu creation. Use v2 endpoint instead.")
	})

	t.Run("should return custom error for HTTP 401", func(t *testing.T) {
		mockResp := &http.Response{}
		mockResp.Status = http.StatusText(http.StatusUnauthorized)
		mockResp.StatusCode = http.StatusUnauthorized
		mockResp.Body = io.NopCloser(strings.NewReader(`{"status": "unauthorized"}`))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.Error(t, err)
		assert.True(t, doguregistry.IsUnauthorizedError(err))
	})

	t.Run("should return custom error for HTTP 403", func(t *testing.T) {
		mockResp := &http.Response{}
		mockResp.Status = http.StatusText(http.StatusForbidden)
		mockResp.StatusCode = http.StatusForbidden
		mockResp.Body = io.NopCloser(strings.NewReader(`{"status": "forbidden"}`))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.Error(t, err)
		assert.True(t, doguregistry.IsForbiddenError(err))
	})

	t.Run("should return custom error for HTTP 404", func(t *testing.T) {
		mockResp := &http.Response{}
		mockResp.Status = http.StatusText(http.StatusNotFound)
		mockResp.StatusCode = http.StatusNotFound
		mockResp.Body = io.NopCloser(strings.NewReader(`{"status": "not found"}`))

		// when
		err := checkStatusCode(mockResp)

		// then
		require.Error(t, err)
		assert.True(t, doguregistry.IsNotFoundError(err))
	})
}

func Test_extractRemoteErrorBody(t *testing.T) {
	t.Run("should return error body", func(t *testing.T) {
		responseBody := io.NopCloser(strings.NewReader(`{"error": "the error text"}`))
		// when
		actual := extractRemoteBody(responseBody, 400)

		// then
		assert.Equal(t, "400: the error text", actual)
	})

	t.Run("should include generic error for truncated json", func(t *testing.T) {
		responseBody := io.NopCloser(strings.NewReader(`{"error": "the erro...`))
		// when
		actual := extractRemoteBody(responseBody, 400)

		// then
		assert.Contains(t, actual, "error while parsing response body")
	})
}

func Test_remoteResponseBody_String(t *testing.T) {
	type fields struct {
		statusCode int
		Status     string
		Error      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"return mixed string", fields{Status: "aaa", Error: "bbb"}, "aaa: bbb"},
		{"return only status", fields{Status: "aaa", Error: ""}, "aaa: (no error)"},
		{"return only error", fields{statusCode: 123, Error: "bbb"}, "123: bbb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseBody := &remoteResponseBody{
				statusCode: tt.fields.statusCode,
				Status:     tt.fields.Status,
				Error:      tt.fields.Error,
			}
			if got := responseBody.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsecureSkipVerify(t *testing.T) {

	// given

	doguNamespace := "official"
	name := "jenkins"
	expectedDogu := createTestDoguV3()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/%s/%s", doguNamespace, name), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedDogu)
	}))
	defer ts.Close()

	// 1. Test with InsecureSkipVerify: false (Should Fail)

	//when
	remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL, InsecureSkipVerify: false}
	dccClient, err := New(remoteConfig, nil)
	require.NoError(t, err)

	//then
	_, err = dccClient.GetLatest(context.Background(), doguNamespace, name)
	require.Error(t, err, "Expected connection to fail with InsecureSkipVerify=false, but it succeeded.")

	// 2. Test with InsecureSkipVerify: true (Should Succeed)

	//when
	remoteConfig = &config.DoguRegistryConfiguration{BaseURL: ts.URL, InsecureSkipVerify: true}
	dccClient, err = New(remoteConfig, nil)
	require.NoError(t, err)

	//then
	actualDogu, err := dccClient.GetLatest(context.Background(), doguNamespace, name)
	require.NoError(t, err, fmt.Sprintf("Expected connection to succeed with InsecureSkipVerify=true, but got error: %v", err))
	assert.Equal(t, expectedDogu, actualDogu)
}
func Test_httpRemote_GetLatest(t *testing.T) {
	doguNamespace := "official"
	name := "jenkins"

	expectedDogu := createTestDoguV3()

	t.Run("should successfully get latest dogu", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/%s/%s", doguNamespace, name), r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDogu)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.GetLatest(context.Background(), doguNamespace, name)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDogu, actualDogu)
	})

	t.Run("should return error for invalid doguNamespace", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.GetLatest(context.Background(), "INVALID_NAMESPACE", name)

		// then
		require.Error(t, err)
		assert.Nil(t, actualDogu)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "namespace of the dogu is not valid")
	})

	t.Run("should return error for invalid dogu name", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.GetLatest(context.Background(), doguNamespace, "INVALID_NAME")

		// then
		require.Error(t, err)
		assert.Nil(t, actualDogu)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "name of the dogu is not valid")
	})

	t.Run("should return not found error when remote returns 404", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.GetLatest(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, actualDogu)
		assert.True(t, doguregistry.IsNotFoundError(err))
	})

	t.Run("should return generic error when parsing response json fails", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.GetLatest(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "failed to parse json of request")
	})
}

func Test_httpRemote_Get(t *testing.T) {
	doguNamespace := "official"
	name := "jenkins"
	version := "0.0.1"

	expectedDogu := createTestDoguV3()

	t.Run("should successfully get dogu version", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/%s/%s/%s", doguNamespace, name, version), r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDogu)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		doguIdentifier := doguv3.Identifier{
			DoguNamespace: doguNamespace,
			Name:          name,
			Version:       version,
		}
		actualDogu, err := dccClient.Get(context.Background(), doguIdentifier)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDogu, actualDogu)
	})

	t.Run("should successfully get dogu version the second time from the cache", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/%s/%s/%s", doguNamespace, name, version), r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDogu)
		}))
		defer ts.Close()

		const CACHE_EXPIRY_SECONDS = 1
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL, DisableCache: false, CacheExpirySeconds: CACHE_EXPIRY_SECONDS}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		doguIdentifier := doguv3.Identifier{
			DoguNamespace: doguNamespace,
			Name:          name,
			Version:       version,
		}
		actualDogu, err := dccClient.Get(context.Background(), doguIdentifier)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDogu, actualDogu)

		stats := dccClient.cache.Stats()
		_, isPresent := dccClient.cache.GetIfPresent(doguIdentifier.String())
		assert.True(t, isPresent)
		assert.Equal(t, uint64(0), stats.Hits)

		//Should get next time from cache
		actualDogu, err = dccClient.Get(context.Background(), doguIdentifier)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDogu, actualDogu)

		_, isPresent = dccClient.cache.GetIfPresent(doguIdentifier.String())
		assert.True(t, isPresent)
		stats = dccClient.cache.Stats()
		//Called once from the Get and twice from the GetIfPresent call in test
		assert.Equal(t, uint64(3), stats.Hits)

		//Should be evicted after designated time
		time.Sleep(time.Duration(CACHE_EXPIRY_SECONDS) * time.Second)

		_, isPresent = dccClient.cache.GetIfPresent(doguIdentifier.String())
		assert.False(t, isPresent)

	})

	t.Run("should return error for invalid doguidentifier", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		tests := []struct {
			name       string
			identifier doguv3.Identifier
		}{
			{
				name:       "invalid dogu namespace",
				identifier: *createTestDoguIdentifer("INVALID", name, version),
			},
			{
				name:       "invalid dogu name",
				identifier: *createTestDoguIdentifer(doguNamespace, "INVALID", version),
			},
			{
				name:       "invalid dogu version",
				identifier: *createTestDoguIdentifer(doguNamespace, name, "INVALID"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				// when
				actualDogu, err := dccClient.Get(context.Background(), tt.identifier)

				// then
				require.Error(t, err)
				assert.Nil(t, actualDogu)
				assert.True(t, doguregistry.IsGenericError(err))
				assert.Contains(t, err.Error(), "dogu identifier is not valid")

			})
		}
	})

	t.Run("should return not found error when remote returns 404", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.Get(context.Background(), *createTestDoguIdentifer(doguNamespace, name, version))

		// then
		require.Error(t, err)
		assert.Nil(t, actualDogu)
		assert.True(t, doguregistry.IsNotFoundError(err))
	})

	t.Run("should return generic error when parsing response json fails", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL + "/"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.Get(context.Background(), *createTestDoguIdentifer(doguNamespace, name, version))

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "failed to parse json of request")
	})
}

func Test_httpRemote_GetVersions(t *testing.T) {
	doguNamespace := "official"
	name := "jenkins"

	t.Run("should successfully return all dogu versions", func(t *testing.T) {
		// given: mock HTTP registry server
		expectedVersions := []string{
			"1.30.4",
			"1.30.3",
			"1.30.1",
		}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/%s/%s/_versions", doguNamespace, name), r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedVersions)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualVersions, err := dccClient.GetVersions(context.Background(), doguNamespace, name)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedVersions, actualVersions)
	})

	t.Run("should return error for invalid doguNamespace", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualVersions, err := dccClient.GetVersions(context.Background(), "INVALID_NAMESPACE", name)

		// then
		require.Error(t, err)
		assert.Nil(t, actualVersions)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "namespace of the dogu is not valid")
	})

	t.Run("should return error for invalid dogu name", func(t *testing.T) {
		// given
		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "http://localhost"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualVersions, err := dccClient.GetVersions(context.Background(), doguNamespace, "INVALID_NAME")

		// then
		require.Error(t, err)
		assert.Nil(t, actualVersions)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "name of the dogu is not valid")
	})

	t.Run("should return not found error when remote returns 404", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		versions, err := dccClient.GetVersions(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, versions)
		assert.True(t, doguregistry.IsNotFoundError(err))
	})

	t.Run("should return generic error when parsing response json fails", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.GetVersions(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "failed to parse response json of request")
	})
}

func Test_httpRemote_GetAll(t *testing.T) {
	doguNamespace := "official"
	name := "jenkins"
	version := "0.0.1"

	name2 := "redmine"
	doguNamespace3 := "testing"

	expectedDoguIdentifier1 := *createTestDoguIdentifer(doguNamespace, name, version)
	expectedDoguIdentifier2 := *createTestDoguIdentifer(doguNamespace, name2, version)
	expectedDoguIdentifier3 := *createTestDoguIdentifer(doguNamespace3, name, version)

	expectedDoguIdentifierList := []doguv3.Identifier{
		expectedDoguIdentifier1,
		expectedDoguIdentifier2,
		expectedDoguIdentifier3,
	}

	t.Run("should successfully get all latest dogu identifiers", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDoguIdentifierList)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)
		actualDoguIdentifiers, err := dccClient.GetAll(context.Background())

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDoguIdentifierList, actualDoguIdentifiers)
	})

	t.Run("should successfully get all latest dogu identifiers with basic auth", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			username, password, ok := r.BasicAuth()
			if !ok || username != "user" || password != "password" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDoguIdentifierList)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		credentials := &config.Credentials{Username: "user", Password: "password"}
		dccClient, err := New(remoteConfig, credentials)
		require.NoError(t, err)
		actualDoguIdentifiers, err := dccClient.GetAll(context.Background())

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedDoguIdentifierList, actualDoguIdentifiers)
	})

	t.Run("should return unauthenticated while getting all latest dogu identifiers with basic auth enabled and wrong user and password", func(t *testing.T) {
		// given: mock HTTP registry server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			username, password, ok := r.BasicAuth()
			if !ok || username != "user" || password != "password" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedDoguIdentifierList)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		credentials := &config.Credentials{Username: "user", Password: "wrongpassword"}
		dccClient, err := New(remoteConfig, credentials)
		require.NoError(t, err)
		actualDoguIdentifiers, err := dccClient.GetAll(context.Background())

		// then
		require.Error(t, err)
		assert.Nil(t, actualDoguIdentifiers)
		assert.True(t, doguregistry.IsUnauthorizedError(err))
	})

	t.Run("should return internal server error when remote returns 500", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		actualDogu, err := dccClient.GetAll(context.Background())

		// then
		require.Error(t, err)
		assert.Nil(t, actualDogu)
		assert.True(t, doguregistry.IsConnectionError(err))
	})

	t.Run("should return generic error when parsing response json fails", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer ts.Close()

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: ts.URL}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.GetAll(context.Background())

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsGenericError(err))
		assert.Contains(t, err.Error(), "failed to parse response json of request")
	})

	t.Run("should successfully return error for wrong protocol", func(t *testing.T) {
		// given:

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "ftp://localhost:8080/"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.GetVersions(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsConnectionError(err))
		assert.Contains(t, err.Error(), "failed to request remote registry")
	})

	t.Run("should successfully return error for wrong protocol", func(t *testing.T) {
		// given:

		remoteConfig := &config.DoguRegistryConfiguration{BaseURL: "ftp://localhost:8080/"}
		dccClient, err := New(remoteConfig, nil)
		require.NoError(t, err)

		// when
		response, err := dccClient.GetVersions(context.Background(), doguNamespace, name)

		// then
		require.Error(t, err)
		assert.Nil(t, response)
		assert.True(t, doguregistry.IsConnectionError(err))
		assert.Contains(t, err.Error(), "failed to request remote registry")
	})

}

func createTestDoguIdentifer(doguNamespace string, name string, version string) *doguv3.Identifier {
	doguIdentifier := &doguv3.Identifier{
		DoguNamespace: doguNamespace,
		Name:          name,
		Version:       version,
	}
	return doguIdentifier
}
func createTestDoguV3() *doguv3.Dogu {
	publishedAt, err := time.Parse(time.RFC3339Nano, "2026-05-06T09:57:04.927Z")
	if err != nil {
		panic(err)
	}

	return &doguv3.Dogu{
		Name:          "redmine",
		DoguNamespace: "official",
		Version:       "0.0.1",
		AppVersion:    "6.1.2",
		PublishedAt:   publishedAt,
		DisplayName:   "Redmine",
		Description:   "Project management and issue tracking",
		Categories:    []string{"Development Apps"},
		Tags: []string{
			"pm",
			"projectmanagement",
			"issue",
			"task",
		},
		Logo:  "https://cloudogu.com/images/dogus/redmine.png",
		URL:   "https://cloudogu.com/ecosystem",
		Chart: "oci://registry.cloudogu.com/official/dogu/v3/charts/redmine",
		Applications: []doguv3.ApplicationVersion{
			{
				Name:    "redmine",
				Version: "6.1.2",
			},
			{
				Name:    "postgresql",
				Version: "16.8",
			},
		},
		Images: []string{
			"registry.cloudogu.com/official/dogu/v3/images/redmine:6.1.2-45.7.0",
			"docker.io/postgres:16.8",
		},
		DoguApis: []string{
			"ServiceAccountRequest.k8s.cloudogu.com/v1",
			"ServiceAccountProducer.k8s.cloudogu.com/v1",
			"Exposition.k8s.cloudogu.com/v1",
			"ConfigValidation.dogu-validation.cloudogu.com/v1",
			"UpgradePath.dogu-migration.cloudogu.com/v1",
		},
		ServiceAccounts: doguv3.ServiceAccounts{
			Requests: []doguv3.ServiceAccountRequest{
				{
					Type:     "nexus",
					Optional: true,
				},
			},
			Producers: []doguv3.ServiceAccountProducer{
				{
					Type: "redmine",
				},
			},
		},
		ExposedPorts: []doguv3.ExposedPort{
			{
				Protocol: "tcp",
				Port:     3000,
			},
		},
		ConfigurableKeys: []string{
			"logging/root",
		},
		Upgrades: []doguv3.Upgrade{
			{
				From:        ">=44.0.0 <45.0.0",
				To:          "45.7.0",
				IsMigration: true,
			},
		},
	}
}
