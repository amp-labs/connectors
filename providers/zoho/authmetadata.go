package zoho

import (
	"context"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho/internal/mail"
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	adapter, err := c.mailAdapterForPostAuth()
	if err != nil {
		slog.Warn("skipping Zoho Mail post-authentication metadata",
			"provider", c.Provider(), "module", c.moduleID, "error", err)

		return &common.PostAuthInfo{}, nil
	}

	resp, accountID, err := adapter.GetAccountID(ctx)
	if err != nil {
		// Expected whenever the connection has no access to Zoho Mail, either
		// because the account has no mailbox or because the Mail scopes were
		// not granted. Account-scoped Mail objects stay unavailable, every
		// other module keeps working.
		slog.Warn("could not resolve Zoho Mail account id, account-scoped Mail objects will be unavailable",
			"provider", c.Provider(), "module", c.moduleID, "error", err)

		return &common.PostAuthInfo{}, nil
	}

	return &common.PostAuthInfo{
		RawResponse: resp,
		CatalogVars: AuthMetadataVars{
			MailAccountID: accountID,
		}.AsMap(),
	}, nil
}

// mailAdapterForPostAuth returns an adapter bound to the Zoho Mail API.
//
// The connector only builds mailAdapter when it was constructed for the Mail
// module. For every other module that field is nil and moduleInfo describes a
// different host, so a dedicated adapter pointed at the Mail base URL is built
// here.
func (c *Connector) mailAdapterForPostAuth() (*mail.Adapter, error) {
	if c.mailAdapter != nil {
		return c.mailAdapter, nil
	}

	return mail.NewAdapter(c.Client, c.providerInfo.ReadModuleInfo(providers.ModuleZohoMail), "")
}
