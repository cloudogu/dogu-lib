package doguv3

import (
	"time"
)

// ApplicationVersion combines a name and version of an application bundled inside a dogu.
type ApplicationVersion struct {
	// Name of the application. As this name is taken from the name of an annotation of the dogu's Helm chart,
	// it must be valid annotation name with no more than 51 characters.
	Name string `json:"Name"`

	// Version is the version number of that application (the format depends on the application)
	Version string `json:"Version"`
}

// ServiceAccountRequest describes a request for a service account a dogu wants to use
// A dogu may request any number of service accounts or none at all.
type ServiceAccountRequest struct {
	// Type is the type (unique name) of the requested service account (e.g. "nexus" or "scm")
	Type string `json:"Type"`

	// Optional states whether the service account is optional or mandatory
	Optional bool `json:"Optional"`
}

// ServiceAccountProducer describes a service account a dogu offers to provide to other dogus.
// In its core it is just a name that can be referenced by other dogus to request a service account of that type.
// A dogu may provide any number of service account types or none at all.
type ServiceAccountProducer struct {
	// Type is the type (unique name) of the provided service account (e.g. "nexus" or "scm")
	Type string `json:"Type"`
}

// ServiceAccounts describes all types of service accounts by which a dogu might communicate with other dogus.
// It can provide service accounts for other dogus, and it can request service accounts from other dogus (where each
// request might be optional or strictly required).
type ServiceAccounts struct {
	// Requests lists all service accounts requested by this dogu (optional or mandatory)
	Requests []ServiceAccountRequest `json:"Requests"`

	// Producers lists all service accounts provided by this dogu
	Producers []ServiceAccountProducer `json:"Producers"`
}

// ExposedPort struct describes a network port a dogu wants to expose. As each port can only exposed by one dogu,
// there might be conflicts, which have to be resolved.
type ExposedPort struct {
	// Protocol the protocol of the exposed port ("tcp" or "udp")
	Protocol string `json:"Protocol"`

	// Port the number of the port the dogu wishes to expose
	Port int `json:"Port"`
}

type Upgrade struct {
	// From specifies the versions this dogu can upgrade from (e.g. ">=44.0.0 <45.0.0")
	From string `json:"From"`

	// To specifies the version number this dogu can upgrade to
	To string `json:"To"`

	// IsMigration specifies whether this upgrade process requires extra (automated) migration steps
	IsMigration bool `json:"IsMigration"`
}

type Dogu struct {

	// The DoguNamespace defines the location where the dogu's chart is provided. It allows to regulate access
	// to dogus in that namespace. There are three reserved dogu namespaces: The namespaces `official` and `k8s`
	// are open to all users without any further costs. Other namespaces (like `premium`) may be restricted to special
	// subscription users, only. The namespace usually consists of
	//   - lower case Latin characters
	//   - special characters underscore "_", minus "-"
	//   - digits 0-9
	//
	// Examples:
	//   - official
	//   - premium
	//   - foo-1
	//

	DoguNamespace string `json:"DoguNamespace"`

	// The Name is used to identify the dogu.
	// The Name must be a DNS compatible identifier and usually consists of
	//   - lower case Latin characters
	//   - special characters underscore "_", minus "-"
	//   - digits 0-9
	// It also should start with a letter (as recommended by Helm).
	//
	// Examples:
	//   - redmine
	//   - confluence
	//   - bar-2
	//
	Name string `json:"Name"`

	// Version is the version of the dogu (which is the version of the dogu's Helm chart
	// and may differ from the application version). The Version must follow the semantic versioning format.
	Version string `json:"Version"`

	// AppVersion is the version of the dogu's (main) application, which may differ from the dogu's Version.
	// It is defined by the "appVersion" field of the dogu's Helm, chart. As the version is defined by the
	// application's developer, its format is not formally restricted.
	AppVersion string `json:"AppVersion"`

	// PublishedAt is the timestamp, when this dogu version has been published.
	// The value is maintained by the dogu registry.
	// Example:
	//   - 2026-05-16T14:57:04.927Z
	PublishedAt time.Time `json:"PublishedAt"`

	// DisplayName is the name of the dogu which is used in UI frontends to represent the dogu. The display name is
	// the value of the Helm chart's "dogu.cloudogu.com/display-name" annotation. It should consist of no more than
	// 30 characters (for practical display reasons), but can use any characters suitable for a UI.
	//
	// Examples:
	//  - Jenkins CI
	//  - Backup & Restore
	//  - SCM-Manager
	DisplayName string `json:"DisplayName"`

	// Description is a (mandatory) short description for the dogu.
	// The description is recommended to consist of a readable sentence which explains shortly the dogu's main topic.
	Description string `json:"Description"`

	// Categories lists the categories under which the dogu should be listed in the Warp menu. Usually this is just one,
	// but dogus are free to declare entries for different categories if it fits their applications.
	// Commonly used categories are "Development Apps", "Administration Apps", or "Documentation", but
	// other categories can be declared as needed.
	Categories []string `json:"Categories"`

	// Tags contains a list of one-word-tags which are in connection with the dogu. This field is optional.
	Tags []string `json:"Tags"`

	// Logo is a URL to an SVG or PNG image suitable to be used as an icon for the dogu. It is defined by the "icon"
	// attribute of the dogu's Helm chart. This field is optional.
	// Example: "https://dogu.cloudogu.com/api/v3/dogus/official/redmine/icon.svg"
	Logo string `json:"Logo"`

	// URL links to the website of the dogu or the original tool vendor. It is defined by the "home" attribute of the
	// dogu's Helm chart. This field is optional.
	// Example: "https://cloudogu.com/ecosystem"
	URL string `json:"URL"`

	// Chart is the OCI reference of the dogu's Helm chart. That Helm chart defines, how the dogu is deployed and
	// integrated into the Cloudogu EcoSystem.
	// Example: "oci://registry.cloudogu.com/official/dogu/v3/charts/redmine"
	Chart string `json:"Chart"`

	// Applications lists one or more application(s) with respective version number bundled by this dogu.
	// Besides the main applications this may contain any helpers like a bundled database.
	Applications []ApplicationVersion `json:"Applications"`

	// Images lists all container images used by this dogu. Each image reference must be fully qualified with
	// registry and version information. This information is extracted from the Helm chart's chart-patch-tpl.yaml.
	// Examples:
	//   - "registry.cloudogu.com/official/dogu/v3/images/redmine:6.1.2-45.7.0",
	//   - "docker.io/postgres:16.8"
	Images []string `json:"Images"`

	// DoguApis lists dogu related APIs and their versions, which are used by this dogu. This information is
	// gathered from the resources of the dogu's Helm chart.
	// Each API is listed in the format <kind>.<group>/<version>.
	// Examples:
	//   - "ServiceAccountRequest.k8s.cloudogu.com/v1"
	//   - "Exposition.k8s.cloudogu.com/v1"
	DoguApis []string `json:"DoguApis"`

	// ServiceAccounts lists all required or provided service accounts for this dogu. This information is collected
	// from the ServiceAccountRequest.k8s.cloudogu.com and ServiceAccountProvider.k8s.cloudogu.com resources contained
	// in the dogu's Helm chart.
	ServiceAccounts ServiceAccounts `json:"ServiceAccounts"`

	// ExposedPorts lists additional ports, this dogu wants to expose. This information
	// is collected from any Exposition.k8s.cloudogu.com resources contained in the dogu's Helm chart.
	// Dogus can expose ports if the dogu provides services to a consumer
	// (f.i. if it wants to provide an API for a CLI tool).
	ExposedPorts []ExposedPort `json:"ExposedPorts"`

	// ConfigurableKeys lists common configuration keys supported by this dogu. The list is taken from the
	// dogu-values-metadata.yaml of the dogu's Helm chart.
	// Example: "logging/root"
	ConfigurableKeys []string `json:"ConfigurableKeys"`

	// Upgrades lists possible upgrade paths this dogu supports. This information is taken from the upgrade-api.yaml
	// of the dogu's Helm chart.
	Upgrades []Upgrade `json:"Upgrades"`
}
