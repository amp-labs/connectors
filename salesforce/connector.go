package salesforce

import (
	"net/http"
	"fmt"
)

type SalesforceConnector struct {
	BaseURL string
	Client *http.Client
}

func NewConnector(workspaceRef string, accessToken string) (*SalesforceConnector, error) {
	return &SalesforceConnector{
		BaseURL: fmt.Sprintf("https://%s.my.salesforce.com/services/data/v52.0", workspaceRef),
		Client: &http.Client{},
	}, nil
}
