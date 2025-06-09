package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/pardot"
)

const (
	apiVersion                 = "59.0"
	versionPrefix              = "v"
	version                    = versionPrefix + apiVersion
	restAPISuffix              = "/services/data/" + version
	uriSobjects                = restAPISuffix + "/sobjects"
	uriToolingEventRelayConfig = restAPISuffix + "/tooling/sobjects/EventRelayConfig"
)

// Connector is a Salesforce connector.
type Connector struct {
	BaseURL   string
	Client    *common.JSONHTTPClient
	XMLClient *common.XMLHTTPClient

	providerInfo  *providers.ProviderInfo
	moduleInfo    *providers.ModuleInfo
	moduleID      common.ModuleID
	pardotAdapter *pardot.Adapter
}

func APIVersionSOAP() string {
	return apiVersion
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parametersInternal{}, opts,
		WithModule(providers.ModuleSalesforceCRM),
	)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		XMLClient: &common.XMLHTTPClient{
			HTTPClient: httpClient,
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.providerInfo, err = providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	conn.setBaseURL(conn.providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: conn.interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: conn.interpretXMLError},
	}.Handle

	// Empty module name, root module, standard salesforce module fallback to default Salesforce behaviour.
	// Account Engagement module will initialize the pardot adapter.
	// Read/Write/ListObjectMetadata will delegate to this adapter.
	moduleID := params.Module.Selection.ID
	if isPardotModule(moduleID) {
		conn.pardotAdapter, err = pardot.NewAdapter(conn.Client, conn.moduleInfo, params.Metadata.Map)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Salesforce
}

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getRestApiURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		restAPISuffix, // scope URLs to API version
	}, paths...)

	return urlbuilder.New(c.BaseURL, parts...)
}

func (c *Connector) getDomainURL(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, paths...)
}

func (c *Connector) getSoapURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, "services/Soap/m", APIVersionSOAP())
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) getURIPartEventRelayConfig(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriToolingEventRelayConfig, paths...)
}

func (c *Connector) getURIPartSobjectsDescribe(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriSobjects, objectName, "describe")
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) isPardotModule() bool {
	return c.pardotAdapter != nil
}

func isPardotModule(moduleID common.ModuleID) bool {
	return moduleID == providers.ModuleSalesforceAccountEngagement ||
		moduleID == providers.ModuleSalesforceAccountEngagementDemo
}

// TODO when new approach to modules is fully done this will be obsolete.
var supportedModules = common.Modules{ // nolint:gochecknoglobals
	providers.ModuleSalesforceCRM: common.Module{
		ID: providers.ModuleSalesforceCRM,
	},
	providers.ModuleSalesforceAccountEngagement: common.Module{
		ID: providers.ModuleSalesforceAccountEngagement,
	},
	providers.ModuleSalesforceAccountEngagementDemo: common.Module{
		ID: providers.ModuleSalesforceAccountEngagementDemo,
	},
}
