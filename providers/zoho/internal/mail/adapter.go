package mail

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

var ErrMissingAccountID = errors.New("missing Zoho Mail account id; post-authentication has not run")

type Adapter struct {
	Client  *common.JSONHTTPClient
	BaseURL string

	// accountID is the Zoho Mail account id (type ZOHO_ACCOUNT) resolved
	// It is required for account-scoped endpoints (e.g.folders, messages)
	accountID string
}

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo, accountID string,
) (*Adapter, error) {
	return &Adapter{
		Client:    client,
		BaseURL:   info.BaseURL,
		accountID: accountID,
	}, nil
}

func (a *Adapter) getAPIURL(path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, path)
}

// getAccountScopedURL builds a URL for an endpoint that lives under a specific
// Zoho Mail account, e.g. api/accounts/{accountId}/messages/view.
func (a *Adapter) getAccountScopedURL(path string) (*urlbuilder.URL, error) {
	if a.accountID == "" {
		return nil, ErrMissingAccountID
	}

	return urlbuilder.New(a.BaseURL, "api/accounts", a.accountID, path)
}
