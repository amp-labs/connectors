package deep

import (
	"errors"
	"fmt"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"github.com/amp-labs/connectors/providers"
)

type Clients struct {
	provider   providers.Provider
	httpClient *common.HTTPClient
	JSON       *common.JSONHTTPClient
	XML        *common.XMLHTTPClient
}

func newClients[P paramsbuilder.ParamAssurance, D dpvars.MetadataVariables](
	provider providers.Provider,
	parameters *dpvars.Parameters[P],
	catalogVars *dpvars.CatalogVariables[P, D],
	errorHandler *interpreter.ErrorHandler,
) (*Clients, error) {
	clientHolder, ok := parameters.Params.(paramsbuilder.ClientHolder)
	if !ok {
		// TODO complain that parameters doesn't hold HTTP client
		return nil, errors.New("not good")
	}

	client := clientHolder.GiveClient().Caller

	providerInfo, err := providers.ReadInfo(provider, catalogVars.List...)
	if err != nil {
		return nil, err
	}

	client.Base = providerInfo.BaseURL

	clients := &Clients{
		provider:   provider,
		httpClient: client,
		JSON: &common.JSONHTTPClient{
			HTTPClient: client,
		},
		XML: &common.XMLHTTPClient{
			HTTPClient: client,
		},
	}

	clients.WithErrorHandler(errorHandler)

	return clients, nil
}

// Provider returns the connector provider.
func (c *Clients) Provider() providers.Provider {
	return c.provider
}

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Clients) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}

// JSONHTTPClient returns the underlying JSON HTTP client.
func (c *Clients) JSONHTTPClient() *common.JSONHTTPClient {
	return c.JSON
}

func (c *Clients) HTTPClient() *common.HTTPClient {
	return c.httpClient
}

func (c *Clients) BaseURL() string {
	return c.httpClient.Base
}

func (c *Clients) WithBaseURL(newURL string) {
	c.httpClient.Base = newURL
}

func (c *Clients) WithErrorHandler(handler *interpreter.ErrorHandler) {
	c.httpClient.ErrorHandler = handler.Handle
}
