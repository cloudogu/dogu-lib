package repository

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudogu/dogu-lib/doguv3"
	"github.com/cloudogu/dogu-lib/doguv3/client/clienterrors"
	"github.com/cloudogu/dogu-lib/doguv3/client/config"

	neturl "net/url"

	"github.com/op/go-logging"
)

// GetLogger is an alias function to provide a different logger for the core.

var Log = logging.MustGetLogger("repository")

func NewRemoteDoguDescriptorRepository(remoteConfig *config.Remote, credentials *config.Credentials) (RemoteDoguDescriptorRepository, error) {
	return newHTTPRemote(remoteConfig, credentials)
}

// httpRemote is able to handle request to a remote registry.
type httpRemote struct {
	endpoint            string
	credentials         *config.Credentials
	client              *http.Client
	remoteConfiguration *config.Remote
}

func newHTTPRemote(remoteConfig *config.Remote, credentials *config.Credentials) (*httpRemote, error) {

	client, err := CreateHTTPClient(remoteConfig)
	if err != nil {
		return nil, err
	}

	return &httpRemote{
		endpoint:            remoteConfig.Endpoint,
		credentials:         credentials,
		client:              client,
		remoteConfiguration: remoteConfig,
	}, nil
}

// CreateHTTPClient creates a httpClient for the given remote settings.
func CreateHTTPClient(config *config.Remote) (*http.Client, error) {
	timeout := 10 * time.Second
	if config.Timeout != 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}
	httpClient := &http.Client{
		Timeout: timeout,
	}

	transport, err := createProxyHTTPTransport(config)
	if err != nil {
		return nil, err
	}
	httpClient.Transport = transport

	return httpClient, nil
}

func createProxyHTTPTransport(config *config.Remote) (*http.Transport, error) {
	transport := &http.Transport{}

	if config.ProxySettings.Enabled {
		proxyURLString := config.ProxySettings.CreateURL()
		Log.Infof("configure http client to use proxy %s", proxyURLString)

		proxyURL, err := neturl.Parse(proxyURLString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy url %s: %w", proxyURLString, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		appendProxyAuthorizationIfRequired(transport, &config.ProxySettings)
	}

	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: config.Insecure,
	}

	return transport, nil
}

func appendProxyAuthorizationIfRequired(transport *http.Transport, proxySettings *config.ProxySettings) {
	if proxySettings.Username != "" {
		authorization := proxySettings.Username + ":" + proxySettings.Password
		basicAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(authorization))
		if transport.ProxyConnectHeader == nil {
			transport.ProxyConnectHeader = make(map[string][]string)
		}

		transport.ProxyConnectHeader.Add("Proxy-Authorization", basicAuthorization)
	}
}

// GetLatest returns the detail about the latest dogu from the remote server by name.
func (r *httpRemote) GetLatest(_ context.Context, doguNamespace string, name string) (*doguv3.Dogu, error) {
	if !doguv3.IsValidName(doguNamespace) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("namespace of the dogu is not valid (doguNamespace: %s)", doguNamespace))
	}
	if !doguv3.IsValidName(name) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("name of the dogu is not valid (name: %s)", name))
	}
	requestUrl := r.endpoint + "/" + doguNamespace + "/" + name
	return r.receiveDoguFromRemoteOrCache(requestUrl)
}

// Get returns a version specific detail about the dogu.
func (r *httpRemote) Get(_ context.Context, doguIdentifier doguv3.Identifier) (*doguv3.Dogu, error) {
	if !doguIdentifier.IsValid() {
		return nil, clienterrors.NewGenericError(fmt.Errorf("dogu identifier is not valid (doguIdentifier: %s)", doguIdentifier.String()))
	}
	requestUrl := r.endpoint + "/" + doguIdentifier.DoguNamespace + "/" + doguIdentifier.Name + "/" + doguIdentifier.Version
	return r.receiveDoguFromRemoteOrCache(requestUrl)
}

// GetVersions returns a version specific dogu descriptor.
func (r *httpRemote) GetVersions(_ context.Context, doguNamespace string, name string) ([]string, error) {
	if !doguv3.IsValidName(doguNamespace) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("namespace of the dogu is not valid (doguNamespace: %s)", doguNamespace))
	}
	if !doguv3.IsValidName(name) {
		return nil, clienterrors.NewGenericError(
			fmt.Errorf("name of the dogu is not valid (name: %s)", name))
	}

	requestURL := r.endpoint + "/" + doguNamespace + "/" + name + "/" + "_versions"
	body, err := r.request(requestURL)
	if err != nil {
		Log.Errorf("failed to request dogu identifiers from remote: %w", err)
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
func (r *httpRemote) GetAll(_ context.Context) ([]doguv3.Identifier, error) {
	requestURL := r.endpoint
	body, err := r.request(requestURL)
	if err != nil {
		Log.Errorf("failed to request dogu identifiers from remote: %w", err)
		return nil, err
	}

	var dogiIdentifiers []doguv3.Identifier
	err = json.Unmarshal(body, &dogiIdentifiers)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse response json of request: %w", err))
	}

	return dogiIdentifiers, nil

}

func (r *httpRemote) receiveDoguFromRemoteOrCache(requestUrl string) (*doguv3.Dogu, error) {
	//TODO caching implementation
	var remoteDogu, err = r.readCachedDogu(requestUrl)
	if err != nil {
		remoteDogu, err = r.requestDogu(requestUrl)

		if err != nil {
			return nil, err
		}

		err = r.writeDoguToCache(remoteDogu, requestUrl)
		if err != nil {
			Log.Errorf("get dogu request was ok but failed to write dogu to cache: %w", err)
		}
	}

	return remoteDogu, nil
}

func (r *httpRemote) readCachedDogu(requestUrl string) (*doguv3.Dogu, error) {
	if r.remoteConfiguration.UseCache {
		//TODO: get the dogu from cache for the version
		/*		cacheFile := filepath.Join(dirname, "content.json")
				doguFromFile, _, err := core.ReadDoguFromFile(cacheFile)
				if err != nil {
					return nil, commonerrors.NewGenericError(fmt.Errorf("failed to read from cache %s: %w", cacheFile, err))
				}
				if doguFromFile == nil {
					return nil, commonerrors.NewNotFoundError(fmt.Errorf("dogu descriptor not found"))
				}
				return doguFromFile, nil*/
	}
	return nil, clienterrors.NewGenericError(fmt.Errorf("useCache is not activated"))
}

func (r *httpRemote) writeDoguToCache(doguToWrite *doguv3.Dogu, requestUrl string) error {
	//TODO: write the dogu to cache for the version

	/*	err := os.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			return commonerrors.NewGenericError(fmt.Errorf("failed to create cache directory %s: %w", dirname, err))
		}

		cacheFile := filepath.Join(dirname, "content.json")
		err = core.WriteDoguToFile(cacheFile, doguToWrite)

		if err != nil {
			removeErr := os.Remove(cacheFile)
			if removeErr != nil {
				core.GetLogger().Warningf("failed to remove cache file %s", cacheFile)
			}
			return commonerrors.NewGenericError(fmt.Errorf("failed to write cache %s: %w", cacheFile, err))
		}*/

	return nil
}

func (r *httpRemote) requestDogu(requestURL string) (*doguv3.Dogu, error) {

	body, err := r.request(requestURL)
	if err != nil {
		Log.Errorf("failed to request dogu from remote: %w", err)
		return nil, err
	}
	var dogu *doguv3.Dogu
	err = json.Unmarshal(body, &dogu)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to parse json of request: %w", err))
	}

	return dogu, nil
}

func (r *httpRemote) request(requestURL string) ([]byte, error) {
	Log.Debugf("fetch json from remote %s", requestURL)

	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, clienterrors.NewGenericError(fmt.Errorf("failed to prepare request: %w", err))
	}

	if r.credentials != nil {
		request.SetBasicAuth(r.credentials.Username, r.credentials.Password)
	}

	resp, err := r.client.Do(request)
	if err != nil {
		return nil, clienterrors.NewConnectionError(fmt.Errorf("failed to request remote registry: %w", err))
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			errClose := resp.Body.Close()
			if errClose != nil {
				Log.Errorf("failed to close body: %w", errClose)
			}
		}
	}()

	err = checkStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
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
	buf := new(strings.Builder)
	_, err := io.Copy(buf, responseBodyReader)
	if err != nil {
		return fmt.Sprintf("error while copying response body: %s", err.Error())
	}

	responseBody := []byte(buf.String())

	body := &remoteResponseBody{statusCode: statusCode}
	jsonErr := json.Unmarshal(responseBody, body)
	if jsonErr != nil {
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
