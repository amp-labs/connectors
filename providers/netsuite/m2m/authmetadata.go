package m2m

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/netsuite"
)

// GetPostAuthInfo retrieves the instance timezone using SuiteQL.
// This is called after authentication to discover instance-specific configuration.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	timezone, err := netsuite.RetrieveInstanceTimezone(ctx, c.ProviderInfo().BaseURL, c.JSONHTTPClient())
	logging.With(ctx, "provider", "netsuiteM2M", "step", "get_post_auth_info")

	isDefault := "false"

	if err != nil {
		timezone, _ = time.LoadLocation(netsuite.DefaultTimezone)
		isDefault = "true"
	}

	c.instanceTimezone = timezone

	catalogVars := map[string]string{
		"sessionTimezone":          timezone.String(),
		"sessionTimezoneIsDefault": isDefault,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}
