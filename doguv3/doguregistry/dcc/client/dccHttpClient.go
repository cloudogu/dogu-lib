package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/config"
	clienterrors "github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/errors"
	"github.com/cloudogu/dogu-lib/doguv3/doguregistry/dcc/logging"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

// dccHttpClient is able to handle request to a remote DCC registry.
type dccHttpClient struct {
	baseURL     *url.URL
	credentials *config.Credentials
	httpClient  *http.Client
	config      *config.DoguRegistryConfiguration
	cache       *otter.Cache[string, *doguv3.Dogu]
}

func New(doguRegistryConfiguration *config.DoguRegistryConfiguration, credentials *config.Credentials) (*dccHttpClient, error) {
	if doguRegistryConfiguration == nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("dogu registry configuration must not be nil"))
	}

	setDefaults(doguRegistryConfiguration)

	httpClient, err := createHTTPClient(doguRegistryConfiguration)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimSuffix(doguRegistryConfiguration.BaseURL, "/")

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse endpoint url %s: %w", endpoint, err))
	}

	dccHttpClient := &dccHttpClient{
		baseURL:     baseURL,
		credentials: credentials,
		httpClient:  httpClient,
		config:      doguRegistryConfiguration,
	}
	if err := initializeCache(doguRegistryConfiguration, dccHttpClient); err != nil {
		slog.Error("failed to initialize cache", "error", err)
		return nil, err
	}

	return dccHttpClient, nil
}

func setDefaults(doguRegistryConfiguration *config.DoguRegistryConfiguration) {
	// Apply defaults when not provided by caller
	if doguRegistryConfiguration.Timeout == 0 {
		doguRegistryConfiguration.Timeout = 10
	}

	if !doguRegistryConfiguration.DisableCache {
		if doguRegistryConfiguration.CacheExpirySeconds == 0 {
			doguRegistryConfiguration.CacheExpirySeconds = 300
		}
		if doguRegistryConfiguration.CacheMaximumDogus == 0 {
			doguRegistryConfiguration.CacheMaximumDogus = 100
		}
	}
}

func initializeCache(doguRegistryConfiguration *config.DoguRegistryConfiguration, doguRegistryClient *dccHttpClient) error {
	if !doguRegistryConfiguration.DisableCache {
		cache, err := otter.New(&otter.Options[string, *doguv3.Dogu]{
			MaximumSize: doguRegistryConfiguration.CacheMaximumDogus,
			ExpiryCalculator: otter.ExpiryAccessing[string, *doguv3.Dogu](
				time.Duration(doguRegistryConfiguration.CacheExpirySeconds) * time.Second), // Reset timer on reads/writes
			StatsRecorder: stats.NewCounter(),
		})
		if err != nil {
			return fmt.Errorf("failed to create otter cache: %w", err)
		}
		doguRegistryClient.cache = cache
	}
	return nil
}

// createHTTPClient creates a http client for the given remote settings.
func createHTTPClient(config *config.DoguRegistryConfiguration) (*http.Client, error) {
	transport, err := createProxyHTTPTransport(config)
	if err != nil {
		slog.Error("Error creating Proxy HttpTransport for DCC Client", "error", err)
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
		Transport: &logging.LoggingRoundTripper{
			Proxied: transport,
		},
	}
	return httpClient, nil
}

func createProxyHTTPTransport(config *config.DoguRegistryConfiguration) (*http.Transport, error) {
	transport := &http.Transport{}

	if config.ProxySettings.Enabled {
		proxyURLString := config.ProxySettings.CreateURL()
		slog.Info("configure http client to use proxy", "proxyURL", proxyURLString)

		proxyURL, err := url.Parse(proxyURLString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy url %s: %w", proxyURLString, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		appendProxyAuthorizationIfRequired(transport, &config.ProxySettings)
	}

	return transport, nil
}

func appendProxyAuthorizationIfRequired(transport *http.Transport, proxySettings *config.ProxySettings) {
	if proxySettings.Username != "" {
		authorization := proxySettings.Username + ":" + proxySettings.Password
		basicAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(authorization))
		if transport.ProxyConnectHeader == nil {
			transport.ProxyConnectHeader = make(http.Header)
		}

		transport.ProxyConnectHeader.Set("Proxy-Authorization", basicAuthorization)
	}
}

// GetLatest returns the detail about the latest dogu from the remote server by name.
func (r *dccHttpClient) GetLatest(ctx context.Context, doguNamespace string, name string) (*doguv3.Dogu, error) {
	if !doguv3.IsValidNamespace(doguNamespace) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("namespace of the dogu is not valid (doguNamespace: %s)", doguNamespace))
	}
	if !doguv3.IsValidName(name) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("name of the dogu is not valid (name: %s)", name))
	}

	requestUrl := r.baseURL.ResolveReference(
		&url.URL{
			Path: path.Join(r.baseURL.Path, doguNamespace, name),
		}).String()
	return r.requestDogu(ctx, requestUrl)
}

// Get returns a version specific detail about the dogu.
func (r *dccHttpClient) Get(ctx context.Context, doguIdentifier doguv3.Identifier) (*doguv3.Dogu, error) {
	if !doguIdentifier.IsValid() {
		return nil, clienterrors.NewGenericError(fmt.Errorf("dogu identifier is not valid (doguIdentifier: %s)", doguIdentifier.String()))
	}

	requestUrl := r.baseURL.ResolveReference(
		&url.URL{
			Path: path.Join(r.baseURL.Path, doguIdentifier.DoguNamespace, doguIdentifier.Name, doguIdentifier.Version),
		}).String()
	return r.requestDoguWithCache(ctx, requestUrl, doguIdentifier)
}

// GetVersions returns a version specific dogu descriptor.
func (r *dccHttpClient) GetVersions(ctx context.Context, doguNamespace string, name string) ([]string, error) {
	if !doguv3.IsValidNamespace(doguNamespace) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("namespace of the dogu is not valid (doguNamespace: %s)", doguNamespace))
	}
	if !doguv3.IsValidName(name) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("name of the dogu is not valid (name: %s)", name))
	}
	versionsPath := "_versions"
	requestURL := r.baseURL.ResolveReference(
		&url.URL{
			Path: path.Join(r.baseURL.Path, doguNamespace, name, versionsPath),
		}).String()

	body, err := r.request(ctx, requestURL)
	if err != nil {
		slog.Error("failed to request dogu identifiers from remote", "error", err)
		return nil, err
	}

	var versions []string
	err = json.Unmarshal(body, &versions)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse response json of request: %w", err))
	}

	return versions, nil

}

// GetAll returns latest doguv3 identifiers of all dogus in the remote server.
func (r *dccHttpClient) GetAll(ctx context.Context) ([]doguv3.Identifier, error) {
	body, err := r.request(ctx, r.baseURL.String())
	if err != nil {
		slog.Error("failed to request dogu identifiers from remote: ", "error", err)
		return nil, err
	}

	var dogiIdentifiers []doguv3.Identifier
	err = json.Unmarshal(body, &dogiIdentifiers)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse response json of request: %w", err))
	}

	return dogiIdentifiers, nil

}

func (r *dccHttpClient) requestDoguWithCache(ctx context.Context, requestUrl string, identifier doguv3.Identifier) (*doguv3.Dogu, error) {
	if !r.config.DisableCache {
		var remoteDogu, doguFound = r.cache.GetIfPresent(identifier.String())
		if doguFound {
			slog.Debug("dogu found in cache", "dogu", remoteDogu)
			return remoteDogu, nil
		}
	}
	remoteDogu, err := r.requestDogu(ctx, requestUrl)

	if err != nil {
		return nil, err
	}

	if !r.config.DisableCache {
		slog.Debug("saving dogu into the cache", "dogu", remoteDogu)
		r.cache.Set(identifier.String(), remoteDogu)
	}

	return remoteDogu, nil
}

func (r *dccHttpClient) requestDogu(ctx context.Context, requestURL string) (*doguv3.Dogu, error) {

	body, err := r.request(ctx, requestURL)
	if err != nil {
		slog.Error("failed to request dogu from remote: ", "error", err)
		return nil, err
	}
	var dogu doguv3.Dogu
	if err = json.Unmarshal(body, &dogu); err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse json of request: %w", err))
	}

	return &dogu, nil
}

func (r *dccHttpClient) request(ctx context.Context, requestURL string) ([]byte, error) {
	slog.Debug("fetch json from remote ", "URL", requestURL)

	request, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to prepare request: %w", err))
	}

	if r.credentials != nil {
		request.SetBasicAuth(r.credentials.Username, r.credentials.Password)
	}

	resp, err := r.httpClient.Do(request)
	if err != nil {
		return nil, clienterrors.NewConnectionError(fmt.Errorf("failed to request remote registry: %w", err))
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			errClose := resp.Body.Close()
			if errClose != nil {
				slog.Error("failed to close body: ", "error", errClose)
			}
		}
	}()

	err = checkStatusCode(resp)
	if err != nil {
		return nil, err
	}

	const maxBodySize = 1 << 23 // 8MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to read response body: %w", err))
	}
	return body, nil
}

func checkStatusCode(response *http.Response) error {
	sc := response.StatusCode
	switch sc {
	case http.StatusUnauthorized:
		return clienterrors.NewUnauthorizedError(errors.New("401 unauthorized, please login to proceed"))
	case http.StatusForbidden:
		return clienterrors.NewForbiddenError(errors.New("403 forbidden, not enough privileges"))
	case http.StatusNotFound:
		return clienterrors.NewNotFoundError(errors.New("404 not found"))
	case http.StatusInternalServerError:
		return clienterrors.NewConnectionError(errors.New("500 internal server error"))
	default:
		if sc >= 300 {
			furtherExplanation := extractRemoteBody(response.Body, sc)

			return clienterrors.NewGenericError(fmt.Errorf("remote registry returns invalid status: %s: %s", response.Status, furtherExplanation))
		}

		return nil
	}
}

func extractRemoteBody(responseBodyReader io.ReadCloser, statusCode int) string {
	const maxBodySize = 1 << 20 // 1MB
	body := &remoteResponseBody{statusCode: statusCode}
	if jsonErr := json.NewDecoder(io.LimitReader(responseBodyReader, maxBodySize)).Decode(body); jsonErr != nil {
		return fmt.Sprintf("error while parsing response body: %s", jsonErr.Error())
	}
	return body.String()
}

type remoteResponseBody struct {
	statusCode int
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func (rb *remoteResponseBody) String() string {
	errorField := rb.Error
	statusField := rb.Status
	if rb.Status == "" {
		statusField = fmt.Sprintf("%d", rb.statusCode)
	}

	if rb.Error == "" {
		errorField = "(no error)"
	}
	return fmt.Sprintf("%s: %s", statusField, errorField)
}
