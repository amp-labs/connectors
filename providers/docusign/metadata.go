package docusign

import (
	"context"
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrNoDefaultAccount = errors.New("no default account found in user info")
	ErrNoAccounts       = errors.New("no accounts found in user info")
	ErrParsingServer    = errors.New("error parsing server from user info")

	userInfoURL = "https://account.docusign.com/oauth/userinfo" // nolint:gochecknoglobals
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) { // nolint:cyclop,funlen
	resp, err := c.get(ctx, userInfoURL)
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, errors.Join(ErrNoAccounts, common.ErrEmptyJSONHTTPResponse)
	}

	var postAuthInfo common.PostAuthInfo
	postAuthInfo.RawResponse = resp

	accounts, err := body.GetKey("accounts")
	if err != nil {
		return nil, err
	}

	array, err := accounts.GetArray()
	if err != nil {
		return nil, err
	}

	if len(array) == 0 {
		return nil, ErrNoAccounts
	}

	for _, account := range array {
		isDefault, err := account.GetKey("is_default")
		if err != nil {
			return nil, err
		}

		val, err := isDefault.GetBool()
		if err != nil {
			return nil, err
		}

		if !val {
			continue
		}

		baseURI, err := account.GetKey("base_uri")
		if err != nil {
			return nil, err
		}

		baseURIString, err := baseURI.GetString()
		if err != nil {
			return nil, err
		}

		if baseURLWithoutHTTPS := strings.TrimPrefix(baseURIString, "https://"); baseURLWithoutHTTPS != baseURIString {
			if parts := strings.SplitN(baseURLWithoutHTTPS, ".", 2); len(parts) > 1 { // nolint:gomnd
				postAuthInfo.CatalogVars = AuthMetadataVars{
					Server: parts[0],
				}.AsMap()

				return &postAuthInfo, nil
			}
		}

		return nil, ErrParsingServer
	}

	return nil, ErrNoDefaultAccount
}
