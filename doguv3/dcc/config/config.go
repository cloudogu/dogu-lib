package config

import (
	"strconv"
)

// DccClientConfiguration contains dogu registry configuration details.
type DccClientConfiguration struct {
	Endpoint           string `validate:"url"`
	ProxySettings      ProxySettings
	Insecure           bool  `json:",omitempty"`
	Timeout            int64 `json:"timeout,omitempty"`
	UseCache           bool  `json:"useCache,omitempty"`
	CacheExpirySeconds int64 `json:"cacheExpirySeconds,omitempty"`
	CacheMaximumDogus  int   `json:"cacheMaximumDogus,omitempty"`
}

// ProxySettings contains the settings for http proxy
type ProxySettings struct {
	Enabled  bool
	Server   string `json:",omitempty"`
	Port     int    `json:",omitempty"`
	Username string `json:",omitempty"`
	Password string `json:",omitempty"`
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
