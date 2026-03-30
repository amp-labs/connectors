// Package okta provides a connector for the Okta Management API.
// API Documentation: https://developer.okta.com/docs/api/
// Users API: https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/
// Groups API: https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/
// Apps API: https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/
package okta

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/okta/metadata"
)

// Connector is the Okta connector.
type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

// NewConnector creates a new Okta connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Okta, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	// Add Reader for read operations
	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	// Add Writer for write operations
	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	// Add Deleter for delete operations
	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}

// ListObjectMetadata returns metadata for the requested objects, including custom fields.
// Custom fields are fetched from the Schema API for users and groups.
// Reference: https://developer.okta.com/docs/reference/api/schemas
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.ProviderContext.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		customFields, err := c.requestCustomFields(ctx, objectName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			continue
		}

		// Ensure maps are initialized to prevent panic when adding fields.
		if objectMetadata.Fields == nil {
			objectMetadata.Fields = make(common.FieldsMetadata)
		}

		if objectMetadata.FieldsMap == nil { //nolint:staticcheck
			objectMetadata.FieldsMap = make(map[string]string) //nolint:staticcheck
		}

		// Add custom fields to object metadata
		for _, field := range customFields {
			// Use the field name as the key (human-readable)
			displayName := field.Title
			if displayName == "" {
				displayName = field.Name
			}

			objectMetadata.AddFieldMetadata(field.Name, common.FieldMetadata{
				DisplayName:  displayName,
				ValueType:    field.getValueType(),
				ProviderType: field.Type,
				Values:       field.getValues(),
			})
		}

		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}
