package zoho

import (
	"context"
	"errors"
	"log"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// zohoAccountType is the account type used to identify the primary Zoho Mail
// mailbox in the /api/accounts response.
const zohoAccountType = "ZOHO_ACCOUNT"

// ErrNoZohoMailAccount is returned when the /api/accounts response contains no
// account of type ZOHO_ACCOUNT.
var ErrNoZohoMailAccount = errors.New("no ZOHO_ACCOUNT found in Zoho Mail accounts response")

// GetPostAuthInfo retrieves post-authentication metadata for the connector.
//
// It is only meaningful for the Zoho Mail module: there it calls the
// /api/accounts endpoint, selects the account whose type is ZOHO_ACCOUNT, and
// returns its accountId as a catalog variable so it can be substituted into the
// account-scoped Zoho Mail API paths.
//
// All other Zoho modules don't require any post-auth call, so an empty
// PostAuthInfo is returned for them.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	ctx = logging.With(ctx, "provider", "zoho", "step", "get_post_auth_info")

	if !c.isMailModule() {
		log.Default().Println("no post-auth info needed for non-mail module")
		return &common.PostAuthInfo{}, nil
	}

	resp, accountID, err := c.retrieveMailAccountID(ctx)
	if err != nil {
		return nil, err
	}

	// Make the resolved id usable on this same connector instance, so callers
	// don't have to rebuild the connector from metadata before listing
	// account-scoped objects.
	c.mailAdapter.SetAccountID(accountID)

	return &common.PostAuthInfo{
		ProviderWorkspaceRef: accountID,
		RawResponse:          resp,
		CatalogVars: AuthMetadataVars{
			MailAccountID: accountID,
		}.AsMap(),
	}, nil
}

func (c *Connector) retrieveMailAccountID(ctx context.Context) (*common.JSONHTTPResponse, string, error) {
	// The connector BaseURL already resolves to the Zoho Mail domain when the
	// Mail module is selected (e.g. https://mail.zoho.com).
	url, err := urlbuilder.New(c.BaseURL, "api/accounts")
	if err != nil {
		return nil, "", err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, "", err
	}

	accountID, err := parseMailAccountID(resp)
	if err != nil {
		return nil, "", err
	}

	return resp, accountID, nil
}

type mailAccountsResponse struct {
	Data []mailAccount `json:"data"`
}

type mailAccount struct {
	Type      string `json:"type"`
	AccountID string `json:"accountId"`
}

func parseMailAccountID(resp *common.JSONHTTPResponse) (string, error) {
	body, err := common.UnmarshalJSON[mailAccountsResponse](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	for _, account := range body.Data {
		if account.Type == zohoAccountType && account.AccountID != "" {
			return account.AccountID, nil
		}
	}

	return "", ErrNoZohoMailAccount
}
