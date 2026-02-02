package phoneburner

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/jsonquery"

	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/phoneburner/metadata"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.PhoneBurner, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	result, err := c.SchemaProvider.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		return nil, err
	}

	// Attach contact custom fields to the contacts object metadata.
	if _, ok := result.Result["contacts"]; ok {
		if err := c.attachContactCustomFieldMetadata(ctx, result); err != nil {
			if result.Errors == nil {
				result.Errors = make(map[string]error)
			}
			result.Errors["contacts"] = errors.Join(common.ErrResolvingCustomFields, err)
		}
	}

	return result, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	return buildReadRequest(ctx, c.ProviderInfo().BaseURL, params)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return parseReadResponse(ctx, params, request, response)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	return buildWriteRequest(ctx, c.ProviderInfo().BaseURL, params)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	return parseWriteResponse(ctx, params, request, response)
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	return buildDeleteRequest(ctx, c.ProviderInfo().BaseURL, params)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return parseDeleteResponse(ctx, params, request, response)
}

func (c *Connector) attachContactCustomFieldMetadata(
	ctx context.Context,
	metadataResult *common.ListObjectMetadataResult,
) error {
	defs, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return err
	}

	objectMetadata := metadataResult.GetObjectMetadata("contacts")
	if objectMetadata == nil {
		return nil
	}

	used := make(map[string]struct{}, len(objectMetadata.Fields))
	for k := range objectMetadata.Fields {
		used[k] = struct{}{}
	}

	for _, def := range defs {
		key := phoneburnerCustomFieldKey(def.DisplayName)
		if key == "" {
			continue
		}

		if _, exists := used[key]; exists {
			key = key + "_" + def.CustomFieldID
		}
		used[key] = struct{}{}

		objectMetadata.AddFieldMetadata(key, common.FieldMetadata{
			DisplayName:  def.DisplayName,
			ValueType:    phoneburnerCustomFieldValueType(def.TypeName),
			ProviderType: def.TypeName,
			IsCustom:     goutils.Pointer(true),
		})
	}

	metadataResult.Result["contacts"] = *objectMetadata

	return nil
}

type phoneburnerCustomFieldDefinition struct {
	CustomFieldID string
	DisplayName   string
	TypeName      string
}

func (c *Connector) fetchCustomFieldDefinitions(ctx context.Context) ([]phoneburnerCustomFieldDefinition, error) {
	// Start from the first page, using the same paging conventions as Read.
	req, err := c.buildReadRequest(ctx, common.ReadParams{ObjectName: "customfields"})
	if err != nil {
		return nil, err
	}

	u, err := urlbuilder.New(req.URL.String())
	if err != nil {
		return nil, err
	}

	// Use the max supported page size for fewer calls.
	u.WithQueryParam("page_size", "100")
	u.WithQueryParam("page", "1")

	next := u.String()
	out := make([]phoneburnerCustomFieldDefinition, 0)

	for next != "" {
		res, err := c.JSONHTTPClient().Get(ctx, next)
		if err != nil {
			return nil, err
		}

		if err := interpretPhoneBurnerEnvelopeError(res); err != nil {
			return nil, err
		}

		node, ok := res.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		records, err := jsonquery.New(node, "customfields").ArrayOptional("customfields")
		if err != nil {
			return nil, err
		}

		for _, r := range records {
			q := jsonquery.New(r)

			customFieldID, err := q.TextWithDefault("custom_field_id", "")
			if err != nil {
				return nil, err
			}
			displayName, err := q.TextWithDefault("display_name", "")
			if err != nil {
				return nil, err
			}
			typeName, err := q.TextWithDefault("type_name", "")
			if err != nil {
				return nil, err
			}

			if customFieldID == "" || displayName == "" {
				continue
			}

			out = append(out, phoneburnerCustomFieldDefinition{
				CustomFieldID: customFieldID,
				DisplayName:   displayName,
				TypeName:      typeName,
			})
		}

		nextPageFunc := nextRecordsURL(next, "customfields")
		next, err = nextPageFunc(node)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
