package zoho

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "crm/v6"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient

	moduleInfo   *providers.ModuleInfo
	providerInfo *providers.ProviderInfo
	moduleID     common.ModuleID

	orgID string
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:          params.Caller.Client,
				ResponseHandler: responseHandler,
			},
		},
		moduleID: params.Module.Selection.ID,
	}

	// Use US region domains as default for testing
	domains, err := GetDomainsForLocation("us")
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(conn.Provider(),
		catalogreplacer.CustomCatalogVariable{
			Plan: catalogreplacer.SubstitutionPlan{
				From: "zoho_api_domain",
				To:   domains.ApiDomain,
			},
		},
		catalogreplacer.CustomCatalogVariable{
			Plan: catalogreplacer.SubstitutionPlan{
				From: "zoho_token_domain",
				To:   domains.TokenDomain,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	conn.setBaseURL(conn.moduleInfo.BaseURL)

	if conn.moduleID == providers.ZohoDeskV2 {
		conn.orgID = conn.
	}

	return conn, nil
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Zoho
}

func (c *Connector) getAPIURL(suffix string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, suffix)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
