package config

import (
	"strconv"
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
