package zoho

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingLocation = errors.New("missing required location in context")
	ErrInvalidLocation = errors.New("invalid location")
)

type LocationDomains struct {
	ApiDomain   string `json:"api_domain"`
	TokenDomain string `json:"token_domain"`
}

func GetDomainsForLocation(location string) (*LocationDomains, error) {
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "":
		return nil, ErrMissingLocation
	case "us":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.com",
			TokenDomain: "accounts.zoho.com",
		}, nil
	case "eu":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.eu",
			TokenDomain: "accounts.zoho.eu",
		}, nil
	case "in":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.in",
			TokenDomain: "accounts.zoho.in",
		}, nil
	case "au":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.com.au",
			TokenDomain: "accounts.zoho.com.au",
		}, nil
	case "cn":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.com.cn",
			TokenDomain: "accounts.zoho.com.cn",
		}, nil
	case "jp":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.jp",
			TokenDomain: "accounts.zoho.jp",
		}, nil
	case "ca":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.ca",
			TokenDomain: "accounts.zohocloud.ca",
		}, nil
	default:
		return nil, fmt.Errorf("%w %q; must be one of US, EU, IN, AU, CN, JP, CA",
			ErrInvalidLocation, location)
	}
}
