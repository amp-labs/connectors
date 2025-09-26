package snapchatads

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// The endpoints shared id(organization Id) in the url path.
var endpointsWithSharedId = datautils.NewSet( //nolint:gochecknoglobals
	"fundingsources",
	"billingcenters",
	"transactions",
	"adaccounts",
	"members",
	"roles",
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	var prefix string

	switch objName {
	case "age_group", "gender", "languages", "advanced_demographics":
		prefix = "targeting/demographics/"
	case "connection_type", "os_type", "carrier", "marketing_name":
		prefix = "targeting/device/"
	case "country":
		prefix = "targeting/geo/"
	case "dlxs", "dlxp", "nln":
		prefix = "targeting/interests/"
	case "categories_loi":
		prefix = "targeting/location/"
	default:
		prefix = ""
	}

	fullPath := prefix + objName

	if endpointsWithSharedId.Has(objName) {
		// If it needs shared ID, build URL with organizationId in the path.
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "organizations", c.organizationId, fullPath)
	}

	// Otherwise, build the normal URL.
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, fullPath)
}
