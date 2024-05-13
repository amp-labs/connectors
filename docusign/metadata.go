package docusign

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNoDefaultAccount = errors.New("no default account found in user info")
	ErrNoAccounts       = errors.New("no accounts found in user info")
	ErrParsingServer    = errors.New("error parsing server from user info")

	userInfoUrl = "https://account.docusign.com/oauth/userinfo"
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (map[string]string, error) {
	resp, err := c.get(ctx, userInfoUrl)
	if err != nil {
		return nil, err
	}

	accounts, err := resp.Body.GetKey("accounts")
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

		if val {
			baseUri, err := account.GetKey("base_uri")
			if err != nil {
				return nil, err
			}

			server, err := baseUri.GetString()
			if err != nil {
				return nil, err
			}

			if baseWithoutHttps := strings.TrimPrefix(server, "https://"); baseWithoutHttps != server {
				if parts := strings.SplitN(baseWithoutHttps, ".", 2); len(parts) > 1 {
					return map[string]string{"server": parts[0]}, nil
				}
			}

			return nil, ErrParsingServer
		}
	}

	return nil, ErrNoDefaultAccount
}
