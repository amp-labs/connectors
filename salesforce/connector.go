package salesforce

import (
	"fmt"
	"net/http"
)

const (
	apiVersion = "v52.0"
)

type SalesforceConnector struct {
	BaseURL     string
	Client      *http.Client
	AccessToken string
}

func NewConnector(workspaceRef string, accessToken string) *SalesforceConnector {
	return &SalesforceConnector{
		BaseURL:     fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s", workspaceRef, apiVersion),
		Client:      &http.Client{},
		AccessToken: accessToken,
	}
}
