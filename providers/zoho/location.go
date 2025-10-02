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
	DeskDomain  string `json:"desk_domain"`
	TokenDomain string `json:"token_domain"`
}

func GetDomainsForLocation(location string) (*LocationDomains, error) {
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "":
		return nil, ErrMissingLocation
	case "us":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.com",
			DeskDomain:  "desk.zoho.com",
			TokenDomain: "accounts.zoho.com",
		}, nil
	case "eu":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.eu",
			DeskDomain:  "desk.zoho.eu",
			TokenDomain: "accounts.zoho.eu",
		}, nil
	case "in":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.in",
			DeskDomain:  "desk.zoho.in",
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
			DeskDomain:  "desk.zoho.com.cn",
			TokenDomain: "accounts.zoho.com.cn",
		}, nil
	case "jp":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.jp",
			DeskDomain:  "desk.zoho.jp",
			TokenDomain: "accounts.zoho.jp",
		}, nil
	case "ca":
		return &LocationDomains{
			ApiDomain:   "www.zohoapis.ca",
			DeskDomain:  "desk.zohocloud.ca",
			TokenDomain: "accounts.zohocloud.ca",
		}, nil
	default:
		return nil, fmt.Errorf("%w %q; must be one of: us, eu, in, au, cn, jp, ca (case-insensitive)",
			ErrInvalidLocation, location)
	}
}
