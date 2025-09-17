package zohocrm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrMissingLocation = errors.New("missing required location in context")
	ErrInvalidLocation = errors.New("invalid location")
)

type locationDomains struct {
	ApiDomain   string `json:"api_domain"`
	TokenDomain string `json:"token_domain"`
}

func locationToDomains(location string) (*locationDomains, error) {
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "us":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.com",
			TokenDomain: "accounts.zoho.com",
		}, nil
	case "eu":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.eu",
			TokenDomain: "accounts.zoho.eu",
		}, nil
	case "in":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.in",
			TokenDomain: "accounts.zoho.in",
		}, nil
	case "au":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.com.au",
			TokenDomain: "accounts.zoho.com.au",
		}, nil
	case "cn":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.com.cn",
			TokenDomain: "accounts.zoho.com.cn",
		}, nil
	case "jp":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.jp",
			TokenDomain: "accounts.zoho.jp",
		}, nil
	case "ca":
		return &locationDomains{
			ApiDomain:   "www.zohoapis.ca",
			TokenDomain: "accounts.zohocloud.ca",
		}, nil
	default:
		return nil, fmt.Errorf("%w %q; must be one of US, EU, IN, AU, CN, JP, CA",
			ErrInvalidLocation, location)
	}
}

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	loc, found := getLocation(ctx)
	if !found {
		return nil, ErrMissingLocation
	}

	domains, err := locationToDomains(loc)
	if err != nil {
		return nil, err
	}

	return &common.PostAuthInfo{
		CatalogVars: &map[string]string{
			"zoho_api_domain":   domains.ApiDomain,
			"zoho_token_domain": domains.TokenDomain,
		},
	}, nil
}
