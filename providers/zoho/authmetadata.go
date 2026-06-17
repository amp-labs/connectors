package zoho

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Modules without a case don't need one, so an empty PostAuthInfo is returned.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	switch c.moduleID {
	case providers.ModuleZohoMail:
		return c.mailPostAuthInfo(ctx)
	default:
		return &common.PostAuthInfo{}, nil
	}
}

// mailPostAuthInfo resolves the Zoho Mail account id (delegated to the mail
// adapter) and returns it as a catalog variable for account-scoped API paths.
func (c *Connector) mailPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	resp, accountID, err := c.mailAdapter.GetAccountID(ctx)
	if err != nil {
		return nil, err
	}

	return &common.PostAuthInfo{
		ProviderWorkspaceRef: accountID,
		RawResponse:          resp,
		CatalogVars: AuthMetadataVars{
			MailAccountID: accountID,
		}.AsMap(),
	}, nil
}
