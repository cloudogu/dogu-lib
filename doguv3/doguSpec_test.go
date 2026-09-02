package doguv3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoguClone(t *testing.T) {
	originalDogu := CreateTestDoguV3()

	clone := originalDogu.Clone()

	assert.Equal(t, originalDogu, clone)
}

func TestDoguCloneNil(t *testing.T) {
	var originalDogu *Dogu

	clone := originalDogu.Clone()

	assert.Equal(t, originalDogu, clone)
}

func CreateTestDoguV3() *Dogu {
	publishedAt, err := time.Parse(time.RFC3339Nano, "2026-05-06T09:57:04.927Z")
	if err != nil {
		panic(err)
	}

	return &Dogu{
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
		Applications: []ApplicationVersion{
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
		ServiceAccounts: ServiceAccounts{
			Requests: []ServiceAccountRequest{
				{
					Type:     "nexus",
					Optional: true,
				},
			},
			Producers: []ServiceAccountProducer{
				{
					Type: "redmine",
				},
			},
		},
		ExposedPorts: []ExposedPort{
			{
				Protocol: "tcp",
				Port:     3000,
			},
		},
		ConfigurableKeys: []string{
			"logging/root",
		},
		Upgrades: []Upgrade{
			{
				From:        ">=44.0.0 <45.0.0",
				To:          "45.7.0",
				IsMigration: true,
			},
		},
	}
}
