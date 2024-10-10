package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion                 = "59.0"
	versionPrefix              = "v"
	version                    = versionPrefix + apiVersion
	restAPISuffix              = "/services/data/" + version
	uriSobjects                = restAPISuffix + "/sobjects"
	uriToolingEventRelayConfig = "tooling/sobjects/EventRelayConfig"
)

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		reader *deep.Reader,
	) *Connector {
		return &Connector{
			Clients:     *clients,
			EmptyCloser: *closer,
			Reader:      *reader,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: interpretXMLError},
	}
	objectURLResolver := deep.SingleURLFormat{
		Produce: func(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			var path string
			switch method {
			case deep.ReadMethod:
				path = "query"
			}

			return urlbuilder.New(baseURL, restAPISuffix, path)
		},
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			// Read reads data from Salesforce. By default, it will read all rows (backfill). However, if Since is set,
			// it will read only rows that have been updated since the specified time.
			// We need to construct the SOQL query.
			url.WithQueryParam("q", makeSOQL(config).String())

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return "records"
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Salesforce, opts,
		errorHandler,
		objectURLResolver,
		firstPage,
		nextPage,
		readObjectLocator,
	)
}

func APIVersionSOAP() string {
	return apiVersion
}

func (c *Connector) getRestApiURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		restAPISuffix, // scope URLs to API version
	}, paths...)

	return urlbuilder.New(c.BaseURL(), parts...)
}

func (c *Connector) getDomainURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL())
}

func (c *Connector) getSoapURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), "services/Soap/m", APIVersionSOAP())
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) getURIPartEventRelayConfig(path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriToolingEventRelayConfig, path)
}

func (c *Connector) getURIPartSobjectsDescribe(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriSobjects, objectName, "describe")
}
