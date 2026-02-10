package bigquery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/amp-labs/connectors/common"
)

// listObjectMetadata retrieves schema information for the specified tables.
// This schema information comes from the Table.Metadata() API, which returns:
//   - Column names and types
//   - Nullability constraints
//   - Nested schema for RECORD types
//   - Partitioning and clustering info (not exposed yet)
//
// BigQuery types are mapped to Ampersand's ValueType enum:
//
//	BigQuery Type    Ampersand ValueType
//	────────────────────────────────────────
//	STRING           String
//	INT64            Integer
//	FLOAT64          Float
//	BOOL             Boolean
//	TIMESTAMP        DateTime
//	DATE             Date
//	BYTES            Binary (not yet mapped)
//	RECORD           Other (nested structures)
//	GEOGRAPHY        Other
//	JSON             Other
//
// BigQuery supports nested/repeated fields via the RECORD type. Currently,
// we mark these as ValueTypeOther.
// TODO: Flatten nested fields.
func (c *Connector) listObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		metadata, err := c.getObjectMetadata(ctx, objectName)
		if err != nil || metadata == nil {
			result.AppendError(objectName, err)

			continue
		}

		result.Result[objectName] = *metadata
	}

	return result, nil
}

// getObjectMetadata retrieves metadata for a single table.
func (c *Connector) getObjectMetadata(ctx context.Context, tableName string) (*common.ObjectMetadata, error) {
	tableRef := c.handle.Dataset(c.dataset).Table(tableName)

	meta, err := tableRef.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata for %s: %w", tableName, err)
	}

	fields := c.schemaToFieldsMetadata(meta.Schema)

	return common.NewObjectMetadata(tableName, fields), nil
}

// schemaToFieldsMetadata converts BigQuery schema to Ampersand FieldsMetadata.
func (c *Connector) schemaToFieldsMetadata(schema []*bigquery.FieldSchema) common.FieldsMetadata {
	fields := make(common.FieldsMetadata)

	for _, field := range schema {
		// Lowercase field names for consistency.
		fieldName := strings.ToLower(field.Name)
		fields[fieldName] = common.FieldMetadata{
			DisplayName:  field.Name, // Keep original casing for display
			ValueType:    bigqueryTypeToValueType(field.Type),
			ProviderType: string(field.Type),
			ReadOnly:     boolPtr(false),
			IsRequired:   boolPtr(field.Required),
		}

		// Handle nested fields (RECORD type)
		if field.Type == bigquery.RecordFieldType && len(field.Schema) > 0 {
			// For nested fields, we could flatten them or handle differently
			// For now, we mark RECORD as other type
			fields[fieldName] = common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    common.ValueTypeOther,
				ProviderType: string(field.Type),
				ReadOnly:     boolPtr(false),
				IsRequired:   boolPtr(field.Required),
			}
		}
	}

	return fields
}
