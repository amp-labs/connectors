package google

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
	Module  common.Module
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts, WithModule(ModuleCalendar))
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		Module: params.Selection,
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(conn.Provider(), getSubdomain(params.Selection.ID))
	if err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	return conn, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module.ID, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, path)
}

// TODO there should be a generic object that accepts declarative object registries.
// Given registries construct URL path. The core URL can be constructed using template and recordID.
// Accept common.Client and use switch case to produce either WriteMethod or DeleteMethod.
//
// DefaultMap still could be relevant.
// This object could be called EndpointLibrary/EndpointCollection.
//
// If we are talking about the read operation this is largely driven by `schema.json`.
// EndpointOperation should be created from schema.json or from similar hard coded EndpointCollection.
//
// Headers/Query parameters can be attached as usual. Payload processing likewise.
//
func (c *Connector) selectEndpointOperation(
	objectName string, recordID string,
) (common.WriteMethod, *urlbuilder.URL, error) {
	registry := supportedObjectsByCreate
	if len(recordID) != 0 {
		registry = supportedObjectsByUpdate
	}

	description := registry[c.Module.ID].Get(objectName)
	if description.IsEmpty() {
		return nil, nil, common.ErrOperationNotSupportedForObject
	}

	url, err := urlbuilder.New(c.BaseURL, description.GetURLPath(recordID))
	if err != nil {
		return nil, nil, err
	}

	switch description.Operation {
	case http.MethodPost:
		return c.Client.Put, url, nil
	case http.MethodPatch:
		return c.Client.Patch, url, nil
	default:
		return nil, nil, common.ErrOperationNotSupportedForObject
	}
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) Provider() providers.Provider {
	return providers.Google
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
