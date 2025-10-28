package snapchatads

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Endpoints that require an organization ID (provided via metadata) as shared ID in the URL path.
var endpointsRequiringOrganizationMetadata = datautils.NewSet( //nolint:gochecknoglobals
	"fundingsources",
	"billingcenters",
	"transactions",
	"adaccounts",
	"members",
	"roles",
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	if endpointsRequiringOrganizationMetadata.Has(objName) {
		// If it needs shared ID, build URL with organizationId in the path.
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "organizations", c.organizationId, objName)
	}

	// Otherwise, build the normal URL.
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objName)
}

func getObjectNodePath(objName string) string {
	// Determine correct node path.
	// For all direct objects, the node path have "targeting_dimensions".
	nodePath := "targeting_dimensions"

	if endpointsRequiringOrganizationMetadata.Has(objName) {
		nodePath = objName
	}

	return nodePath
}
