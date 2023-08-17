package salesforce

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

const (
	apiVersion = "v58.0"
)

// Connector is a Salesforce connector.
type Connector struct {
	Domain      string
	BaseURL     string
	Client      *http.Client
	AccessToken func() (string, error)
}

// NewConnector returns a new Salesforce connector.
func NewConnector(workspaceRef string, getToken common.TokenProvider[string]) *Connector {
	return &Connector{
		BaseURL:     fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s", workspaceRef, apiVersion),
		Domain:      fmt.Sprintf("%s.my.salesforce.com", workspaceRef),
		Client:      http.DefaultClient,
		AccessToken: getToken,
	}
}
