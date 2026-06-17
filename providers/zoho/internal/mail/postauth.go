package mail

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
)

// zohoAccountType is the account type used to identify the primary Zoho Mail
// mailbox in the /api/accounts response.
const zohoAccountType = "ZOHO_ACCOUNT"

var ErrNoZohoMailAccount = errors.New("no ZOHO_ACCOUNT found in Zoho Mail accounts response")

// GetAccountID fetches the Zoho Mail accounts, selects the ZOHO_ACCOUNT
// entry, stores its id on the adapter, and returns the raw response and id.
func (a *Adapter) GetAccountID(ctx context.Context) (*common.JSONHTTPResponse, string, error) {
	url, err := a.getAPIURL("api/accounts")
	if err != nil {
		return nil, "", err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, "", err
	}

	accountID, err := parseMailAccountID(resp)
	if err != nil {
		return nil, "", err
	}

	a.accountID = accountID

	return resp, accountID, nil
}

type accountsResponse struct {
	Data []account `json:"data"`
}

type account struct {
	Type      string `json:"type"`
	AccountID string `json:"accountId"`
}

func parseMailAccountID(resp *common.JSONHTTPResponse) (string, error) {
	body, err := common.UnmarshalJSON[accountsResponse](resp)
	if err != nil {
		return "", common.ErrFailedToUnmarshalBody
	}

	for _, acc := range body.Data {
		if acc.Type == zohoAccountType && acc.AccountID != "" {
			return acc.AccountID, nil
		}
	}

	return "", ErrNoZohoMailAccount
}
