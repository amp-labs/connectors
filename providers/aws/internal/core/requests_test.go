package core

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/providers"
	"github.com/stretchr/testify/require"
)

// Tests that domain is correctly replaced.
// Domain is such variable that changes based on the request, and it is not scoped to the connector level.
func TestBuildRequest(t *testing.T) {
	t.Parallel()

	regionVariable := catalogreplacer.CustomCatalogVariable{
		Plan: catalogreplacer.SubstitutionPlan{
			From: "region",
			To:   "test-region",
		},
	}

	info, err := providers.ReadInfo(providers.AWS, regionVariable)
	require.NoError(t, err)

	module, err := info.ReadModuleInfo(providers.ModuleAWSIdentityCenter, regionVariable)
	require.NoError(t, err)

	baseURL := module.BaseURL

	request, err := BuildRequest(context.Background(), baseURL,
		"test-domain", "test-service", "test-command", nil)
	require.NoError(t, err)

	actual := request.URL.String()

	require.Equal(t, "https://test-domain.test-region.amazonaws.com", actual)
}
