package config

import (
	"strconv"
)

type RetryPolicy struct {
	Type          string `json:"type,omitempty"`
	Interval      int64  `json:"interval" validate:"gte=0"`
	MaxRetryCount int    `json:"maxRetryCount" validate:"gte=0"`
}

// Registry contains Cloudogu EcoSystem registration details.
type Registry struct {
	Type        string      `validate:"eq=etcd"`
	Endpoints   []string    `validate:"required,min=1"`
	RetryPolicy RetryPolicy `json:"retryPolicy,omitempty"`
}

// Remote contains dogu registry configuration details.
type Remote struct {
	Endpoint               string `validate:"url"`
	AuthenticationEndpoint string `validate:"omitempty,url"`
	ProxySettings          ProxySettings
	AnonymousAccess        bool        `json:",omitempty"`
	Insecure               bool        `json:",omitempty"`
	RetryPolicy            RetryPolicy `json:"retryPolicy,omitempty"`
	Timeout                int64       `json:"timeout,omitempty"`
	UseCache               bool        `json:"useCache,omitempty"`
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
