package callrail

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	accountEndpoint = "v3/a.json"
)

type AccountsResponse struct {
	Accounts []map[string]any `json:"accounts"`
}

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	account, rawResp, err := c.retrieveAccountId(ctx)
	if err != nil {
		return nil, err
	}

	c.accountId = account

	cv := map[string]string{
		"account_id": account,
	}

	return &common.PostAuthInfo{
		CatalogVars: &cv,
		RawResponse: rawResp,
	}, nil
}

func (c *Connector) retrieveAccountId(ctx context.Context) (string, *common.JSONHTTPResponse, error) {
	fullURL, err := urlbuilder.New(c.ModuleInfo().BaseURL, accountEndpoint)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build user info URL: %w", err)
	}

	resp, err := c.JSONHTTPClient().Get(ctx, fullURL.String())
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	res, err := common.UnmarshalJSON[AccountsResponse](resp)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse account info response: %w", err)
	}

	if len(res.Accounts) == 0 {
		return "", nil, common.ErrMissingExpectedValues
	}

	accountId, ok := res.Accounts[0]["numeric_id"].(float64)
	if !ok {
		return "", nil, common.ErrMissingExpectedValues
	}

	return strconv.Itoa(int(accountId)), resp, nil
}
