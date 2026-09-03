package config

import (
	"strconv"
)

const (
	// URLSchemaDefault addresses a live DCC-v3 API, where every route is a resource of its own. This is the
	// default when no schema is configured.
	URLSchemaDefault = "default"
	// URLSchemaIndex addresses a file-based DCC, that is a DCC mirrored onto a static webserver. Routes that are
	// a directory there are answered by a file inside that directory.
	URLSchemaIndex = "index"
)

// DoguRegistryConfiguration contains dogu registry configuration details.
type DoguRegistryConfiguration struct {
	BaseURL            string
	ProxySettings      ProxySettings
	Timeout            int64
	DisableCache       bool
	CacheExpirySeconds int64
	CacheMaximumDogus  int
	InsecureSkipVerify bool
	UserAgent          string
	URLSchema          string
}

// ProxySettings contains the settings for http proxy
type ProxySettings struct {
	Enabled  bool
	Server   string
	Port     int
	Username string
	Password string
}

// CreateURL creates a proxy http url
func (proxy ProxySettings) CreateURL() string {
	return "http://" + proxy.Server + ":" + strconv.Itoa(proxy.Port)
}

// Credentials for a remote system
type Credentials struct {
	Username string
	Password string
}
