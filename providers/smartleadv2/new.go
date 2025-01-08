// nolint
package smartleadv2

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/del"
	"github.com/amp-labs/connectors/internal/components/metadata"
	"github.com/amp-labs/connectors/internal/components/read"
	"github.com/amp-labs/connectors/internal/components/write"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/spyzhov/ajson"
)

// This reads the OpenAPI schema from the embedded file.
var (
	//go:embed schema.json
	smartleadSchemas []byte

	FileManager = scrapper.NewMetadataFileManager(smartleadSchemas, fileconv.NewSiblingFileLocator()) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)

// TODO: Could this be a part of the providerInfo object?
const (
	apiVersion             = "v1"
	objectNameCampaign     = "campaigns"
	objectNameEmailAccount = "email-accounts"
	objectNameClient       = "client"
)

// Connector is the Smartlead connector.
type Connector struct {
	components.ConnectorComponent

	// Input parameters that are required for the connector to function.
	common.RequireAuthenticatedClient
}

func NewConnector(params common.Parameters) (conn *Connector, outErr error) {
	s, err := support()
	if err != nil {
		return nil, err
	}

	return components.Initialize(
		providers.Smartlead,
		params,
		setup,
		components.WithProviderEndpointSupport(*s),
		components.WithErrorHandler(interpreter.ErrorHandler{
			JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
		}),
	)
}

// setup is a constructor for the Smartlead connector.
func setup(connectorComponent *components.ConnectorComponent) (*Connector, error) {
	conn := &Connector{
		ConnectorComponent: *connectorComponent,
	}

	conn.MetadataStrategy = metadata.NewMultipleStrategy(
		metadata.NewSingleObjectEndpointStrategy(conn.JSON.HTTPClient.Client, metadataSampleRequestBuilder(conn), metadataSampleResponseParser(conn)),
		metadata.NewOpenAPIStrategy(Schemas, conn.Module()),
	)

	conn.ReadStrategy = read.NewSimpleReadStrategy(
		conn.JSON.HTTPClient.Client,
		readRequestBuilder(conn),
		readResponseParser(),
	)

	conn.WriteStrategy = write.NewSimpleWriteStrategy(
		conn.JSON.HTTPClient.Client,
		createRequestBuilder(conn),
		updateRequestBuilder(conn),
		writeResponseParser(),
	)

	conn.DeleteStrategy = del.NewSimpleDeleteStrategy(
		conn.JSON.HTTPClient.Client,
		deleteRequestBuilder(conn),
	)

	// Behavior overrides can go in here.
	return conn, nil
}

// support returns the currently supported endpoints for the Smartlead connector.
func support() (*components.ProviderEndpointSupport, error) {
	a := map[common.ModuleID][]components.EndpointSupport{
		"": {
			{
				Endpoint: "{campaigns,client,leads,email-accounts}",
				Support:  components.ReadSupport,
			},
			{
				Endpoint: "{campaigns}",
				Support:  components.DeleteSupport,
			},
			{
				Endpoint: "{campaigns,client,email-accounts}",
				Support:  components.WriteSupport,
			},

			// Everything else is rejected.
		},
	}

	return components.NewProviderEndpointSupport(a)
}

// metadataSampleRequestBuilder makes a request to sample an object.
// TODO: What happens when response is empty?
func metadataSampleRequestBuilder(conn *Connector) metadata.RequestBuilder {
	return func(ctx context.Context, object string) (*http.Request, error) {
		return buildRequest(ctx, http.MethodGet, nil, conn.ProviderInfo().BaseURL, apiVersion, object)
	}
}

// metadataSampleResponseParser parses the response to get metadata.
func metadataSampleResponseParser(conn *Connector) metadata.ResponseParser {
	return func(ctx context.Context, response *http.Response) (*common.ObjectMetadata, error) {
		result, err := parseResponse[[]map[string]any](response.Body)
		if err != nil {
			return nil, err
		}

		if len(*result) == 0 {
			return nil, fmt.Errorf("no metadata found")
		}

		// Get the keys of the first map
		m := common.ObjectMetadata{
			FieldsMap:   make(map[string]string),
			DisplayName: "",
		}

		for k := range (*result)[0] {
			m.FieldsMap[k] = k
		}

		return &m, nil
	}
}

// readRequestBuilder makes a request to read an object.
func readRequestBuilder(conn *Connector) read.RequestBuilder {
	return func(ctx context.Context, config common.ReadParams) (*http.Request, error) {
		return buildRequest(ctx, http.MethodGet, nil, conn.ProviderInfo().BaseURL, apiVersion, config.ObjectName)
	}
}

// readResponseParser parses the response to get the read result.
func readResponseParser() read.ResponseParser {
	return func(ctx context.Context, params common.ReadParams, response *http.Response) (*common.ReadResult, error) {
		jsonResponse, err := common.ParseJSONResponse(response)
		if err != nil {
			return nil, err
		}

		return common.ParseResult(
			jsonResponse,
			getRecords,
			getNextRecordsURL,
			common.GetMarshaledData,
			params.Fields,
		)
	}
}

func createRequestBuilder(conn *Connector) write.RequestBuilder {
	return func(ctx context.Context, config common.WriteParams) (*http.Request, error) {
		var url, suffix = []string{conn.ProviderInfo().BaseURL, apiVersion, config.ObjectName}, ""

		switch config.ObjectName {
		case objectNameCampaign:
			suffix = "create"
		case objectNameEmailAccount, objectNameClient:
			suffix = "save"
		}

		return buildRequest(ctx, http.MethodPost, config.RecordData, append(url, suffix)...)
	}
}

func updateRequestBuilder(conn *Connector) write.RequestBuilder {
	return func(ctx context.Context, config common.WriteParams) (*http.Request, error) {
		var url = []string{conn.ProviderInfo().BaseURL, apiVersion, config.ObjectName}

		if config.ObjectName == objectNameEmailAccount {
			url = append(url, config.RecordId)
		}

		return buildRequest(ctx, http.MethodPost, config.RecordData, url...)
	}
}

func writeResponseParser() write.ResponseParser {
	return func(ctx context.Context, params common.WriteParams, response *http.Response) (*common.WriteResult, error) {
		b, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		root, err := ajson.Unmarshal(b)
		if err != nil {
			return nil, err
		}

		if root == nil {
			return &common.WriteResult{
				Success: true,
			}, nil
		}

		var recordIdPaths = map[string]string{ // nolint:gochecknoglobals
			objectNameCampaign:     "id",
			objectNameEmailAccount: "emailAccountId",
			objectNameClient:       "clientId",
		}

		recordIdNodePath := recordIdPaths[params.ObjectName]

		// ID is integer that is always stored under different field name.
		rawID, err := jsonquery.New(root).Integer(recordIdNodePath, true)
		if err != nil {
			return nil, err
		}

		recordID := ""
		if rawID != nil {
			// optional
			recordID = strconv.FormatInt(*rawID, 10)
		}

		return &common.WriteResult{
			Success:  true,
			RecordId: recordID,
			Errors:   nil,
			Data:     nil,
		}, nil
	}
}

func deleteRequestBuilder(conn *Connector) del.RequestBuilder {
	return func(ctx context.Context, config common.DeleteParams) (*http.Request, error) {
		return buildRequest(
			ctx,
			http.MethodDelete,
			nil,
			conn.ProviderInfo().BaseURL,
			apiVersion,
			config.ObjectName,
			config.RecordId,
		)
	}
}

// TODO: Could be moved to common
func buildRequest(ctx context.Context, method string, body any, urls ...string) (*http.Request, error) {
	url, err := urlbuilder.New(urls[0], urls[1:]...)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(b))
}

func parseResponse[T any](body io.Reader) (*T, error) {
	var result T
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).Array("", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func getNextRecordsURL(_ *ajson.Node) (string, error) {
	// Pagination is not supported for this provider.
	return "", nil
}
