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
	ApiDomain             string `json:"api_domain"`
	DeskDomain            string `json:"desk_domain"`
	ServiceDeskPlusDomain string `json:"servicedesk_plus_domain"`
	TokenDomain           string `json:"token_domain"`
}

func GetDomainsForLocation(location string) (*LocationDomains, error) { // nolint: cyclop,funlen
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "":
		return nil, ErrMissingLocation
	case "us":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.com",
			DeskDomain:            "desk.zoho.com",
			ServiceDeskPlusDomain: "sdpondemand.manageengine.com",
			TokenDomain:           "accounts.zoho.com",
		}, nil
	case "eu":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.eu",
			DeskDomain:            "desk.zoho.eu",
			ServiceDeskPlusDomain: "sdpondemand.manageengine.eu",
			TokenDomain:           "accounts.zoho.eu",
		}, nil
	case "in":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.in",
			DeskDomain:            "desk.zoho.in",
			ServiceDeskPlusDomain: "sdpondemand.manageengine.in",
			TokenDomain:           "accounts.zoho.in",
		}, nil
	case "au":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.com.au",
			ServiceDeskPlusDomain: "servicedeskplus.net.au",
			TokenDomain:           "accounts.zoho.com.au",
		}, nil
	case "cn":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.com.cn",
			DeskDomain:            "desk.zoho.com.cn",
			ServiceDeskPlusDomain: "servicedeskplus.cn",
			TokenDomain:           "accounts.zoho.com.cn",
		}, nil
	case "jp":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.jp",
			DeskDomain:            "desk.zoho.jp",
			ServiceDeskPlusDomain: "servicedeskplus.jp",
			TokenDomain:           "accounts.zoho.jp",
		}, nil
	case "ca":
		return &LocationDomains{
			ApiDomain:             "www.zohoapis.ca",
			DeskDomain:            "desk.zohocloud.ca",
			ServiceDeskPlusDomain: "servicedeskplus.ca",
			TokenDomain:           "accounts.zohocloud.ca",
		}, nil
	case "uk":
		return &LocationDomains{
			ServiceDeskPlusDomain: "servicedeskplus.uk",
			TokenDomain:           "accounts.zoho.sa",
		}, nil
	case "sa":
		return &LocationDomains{
			ServiceDeskPlusDomain: "servicedeskplus.sa",
			TokenDomain:           "accounts.zoho.uk",
		}, nil
	default:
		return nil, fmt.Errorf("%w %q; must be one of: us, eu, in, au, cn, jp, ca, uk, sa (case-insensitive)",
			ErrInvalidLocation, location)
	}
}
