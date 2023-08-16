package salesforce

import (
	"fmt"
	"net/http"
)

const (
	apiVersion = "v58.0"
)

type Connector struct {
	Domain      string
	BaseURL     string
	Client      *http.Client
	AccessToken func() (string, error)
}

func NewConnector(workspaceRef string, getToken func() (string, error)) *Connector {
	return &Connector{
		BaseURL:     fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s", workspaceRef, apiVersion),
		Domain:      fmt.Sprintf("%s.my.salesforce.com", workspaceRef),
		Client:      http.DefaultClient,
		AccessToken: getToken,
	}
}
