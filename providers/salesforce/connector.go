package salesforce

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
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
	deep.Writer
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
		writer *deep.Writer,
	) *Connector {
		return &Connector{
			Clients:     *clients,
			EmptyCloser: *closer,
			Reader:      *reader,
			Writer:      *writer,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: interpretXMLError},
	}
	objectURLResolver := dpobjects.URLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			switch method {
			case dpobjects.ReadMethod:
				return urlbuilder.New(baseURL, restAPISuffix, "query")
			case dpobjects.CreateMethod:
				return urlbuilder.New(baseURL, restAPISuffix, "sobjects", objectName)
			case dpobjects.UpdateMethod:
				url, err := urlbuilder.New(baseURL, restAPISuffix, "sobjects", objectName)
				if err != nil {
					return nil, err
				}
				// Salesforce allows PATCH method which will act as Update.
				url.WithQueryParam("_HttpMethod", "PATCH")

				return url, nil
			}

			// TODO general error
			return nil, errors.New("cannot match URL for object")
		},
	}
	firstPage := dpread.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			// Read reads data from Salesforce. By default, it will read all rows (backfill). However, if Since is set,
			// it will read only rows that have been updated since the specified time.
			// We need to construct the SOQL query.
			url.WithQueryParam("q", makeSOQL(config).String())

			return url, nil
		},
	}
	nextPage := dpread.NextPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
		},
	}
	readObjectLocator := dpread.ResponseLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return "records"
		},
	}
	writeResultBuilder := dpwrite.ResponseBuilder{
		Build: writeResultBuild,
	}

	return deep.Connector[Connector, parameters](constructor, providers.Salesforce, opts,
		errorHandler,
		objectURLResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		dpwrite.PostPostWriteRequestBuilder{},
		writeResultBuilder,
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

func writeResultBuild(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(body).Str("id", false)
	if err != nil {
		return nil, err
	}

	errorsList, err := getErrors(body)
	if err != nil {
		return nil, err
	}

	success, err := jsonquery.New(body).Bool("success", false)
	if err != nil {
		return nil, err
	}

	// Salesforce does not return record data upon successful write so we do not populate
	// the corresponding result field
	return &common.WriteResult{
		RecordId: *recordID,
		Errors:   errorsList,
		Success:  *success,
	}, nil
}

// getErrors returns the errors from the response.
func getErrors(node *ajson.Node) ([]any, error) {
	arr, err := jsonquery.New(node).Array("errors", true)
	if err != nil {
		return nil, err
	}

	objects, err := jsonquery.Convertor.ArrayToObjects(arr)
	if err != nil {
		return nil, err
	}

	return objects, nil
}
