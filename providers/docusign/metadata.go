package docusign

import (
	"context"
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/docusign/metadata"
	"github.com/spyzhov/ajson"
)

var (
	ErrNoDefaultAccount = errors.New("no default account found in user info")
	ErrNoAccounts       = errors.New("no accounts found in user info")
	ErrParsingServer    = errors.New("error parsing server from user info")

	userInfoURL = "https://account.docusign.com/oauth/userinfo" // nolint:gochecknoglobals
)

//nolint:funlen,cyclop
func (c *Connector) GetPostAuthInfo(
	ctx context.Context,
) (*common.PostAuthInfo, error) { // nolint:cyclop,funlen
	resp, err := c.get(ctx, userInfoURL)
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, errors.Join(ErrNoAccounts, common.ErrEmptyJSONHTTPResponse)
	}

	postAuthInfo := common.PostAuthInfo{
		RawResponse: resp,
	}

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

		c.accountId = getAccountId(account)
		authVars := AuthMetadataVars{
			AccountId: c.accountId,
		}

		baseURI, err := account.GetKey("base_uri")
		if err != nil {
			return nil, err
		}

		baseURIString, err := baseURI.GetString()
		if err != nil {
			return nil, err
		}

		if baseURLWithoutHTTPS, found := strings.CutPrefix(baseURIString, "https://"); found {
			if parts := strings.SplitN(baseURLWithoutHTTPS, ".", 2); len(parts) > 1 { // nolint:mnd
				authVars.Server = parts[0]
				postAuthInfo.CatalogVars = authVars.AsMap()

				return &postAuthInfo, nil
			}
		}

		return nil, ErrParsingServer
	}

	return nil, ErrNoDefaultAccount
}

func getAccountId(account *ajson.Node) string {
	accountId, err := account.GetKey("account_id")
	if err != nil {
		// The connector can still be used as a proxy if we've failed to retrieve the account ID.
		return ""
	}

	accountIdString, err := accountId.GetString()
	if err != nil {
		return ""
	}

	return accountIdString
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Schemas.Select(common.ModuleRoot, objectNames)
}
