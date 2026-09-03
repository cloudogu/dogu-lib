package testutil

import (
	"time"

	"github.com/cloudogu/dogu-lib/doguv3"
)

func CreateTestDoguV3() *doguv3.Dogu {
	publishedAt, err := time.Parse(time.RFC3339Nano, "2026-05-06T09:57:04.927Z")
	if err != nil {
		panic(err)
	}

	testDoguV3 := &doguv3.Dogu{
		Name:          "redmine",
		DoguNamespace: "official",
		Version:       "0.0.1",
		AppVersion:    "6.1.2",
		PublishedAt:   publishedAt,
		DisplayName:   "Redmine",
		Description:   "Project management and issue tracking",

		Logo:  "https://cloudogu.com/images/dogus/redmine.png",
		URL:   "https://cloudogu.com/ecosystem",
		Chart: "oci://registry.cloudogu.com/official/dogu/v3/charts/redmine",
	}
	appendStringArrayParameters(testDoguV3)
	appendedApplicationVersions(testDoguV3)
	appendServiceAccounts(testDoguV3)
	appendExportPorts(testDoguV3)
	appendUpgrades(testDoguV3)

	return testDoguV3
}

func appendUpgrades(testDoguV3 *doguv3.Dogu) {
	testDoguV3.Upgrades = []doguv3.Upgrade{
		{
			From:        ">=44.0.0 <45.0.0",
			To:          "45.7.0",
			IsMigration: true,
		},
	}
}

func appendExportPorts(testDoguV3 *doguv3.Dogu) {

	testDoguV3.ExposedPorts = []doguv3.ExposedPort{
		{
			Protocol: "tcp",
			Port:     3000,
		},
	}

}

func appendServiceAccounts(testDoguV3 *doguv3.Dogu) {

	testDoguV3.ServiceAccounts = doguv3.ServiceAccounts{
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
	}
}

func appendedApplicationVersions(testDoguV3 *doguv3.Dogu) {
	testDoguV3.Applications = []doguv3.ApplicationVersion{
		{
			Name:    "redmine",
			Version: "6.1.2",
		},
		{
			Name:    "postgresql",
			Version: "16.8",
		},
	}
}

func appendStringArrayParameters(testDoguV3 *doguv3.Dogu) {

	testDoguV3.Categories = []string{"Development Apps"}
	testDoguV3.Tags = []string{
		"pm",
		"projectmanagement",
		"issue",
		"task",
	}

	testDoguV3.Images = []string{
		"registry.cloudogu.com/official/dogu/v3/images/redmine:6.1.2-45.7.0",
		"docker.io/postgres:16.8",
	}
	testDoguV3.DoguApis = []string{
		"ServiceAccountRequest.k8s.cloudogu.com/v1",
		"ServiceAccountProducer.k8s.cloudogu.com/v1",
		"Exposition.k8s.cloudogu.com/v1",
		"ConfigValidation.dogu-validation.cloudogu.com/v1",
		"UpgradePath.dogu-migration.cloudogu.com/v1",
	}

	testDoguV3.ConfigurableKeys = []string{
		"logging/root",
	}
}
