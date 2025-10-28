package zoho

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	crmAPIVersion   = "crm/v6"
	deskAPIVersion  = "api/v1"
	defaultLocation = "us"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient

	moduleInfo   *providers.ModuleInfo
	providerInfo *providers.ProviderInfo
	moduleID     common.ModuleID
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ZohoCRM), // The module is resolved on behalf of the user if the option is missing.
		WithLocation(defaultLocation), // Use US region as default for testing
	)
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

	var domains *LocationDomains

	if params.domains == nil {
		domains, err = GetDomainsForLocation(params.location)
		if err != nil {
			return nil, err
		}
	} else {
		domains = params.domains
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
				From: "zoho_desk_domain",
				To:   domains.DeskDomain,
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

	conn.providerInfo = providerInfo
	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)
	conn.setBaseURL(conn.moduleInfo.BaseURL)

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

func (c *Connector) getAPIURL(apiVersion string, suffix string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, suffix)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
