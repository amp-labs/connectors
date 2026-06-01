package monday

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// listObjectMetadata returns GraphQL introspection metadata and merges board column
// definitions into items when board_id is provided via items@<board_id> object name.
func (c *Connector) listObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	result := common.NewListObjectMetadataResult()

	for _, rawObjectName := range objectNames {
		objectName, boardID := parseObjectNameAndBoardID(rawObjectName)

		objectMetadata, err := c.fetchObjectMetadataViaIntrospection(ctx, objectName)
		if err != nil {
			result.AppendError(rawObjectName, err)

			continue
		}

		if objectsWithCustomFields.Has(objectName) && boardID != "" {
			columns, fetchErr := c.fetchBoardColumnDefinitions(ctx, boardID)
			if fetchErr != nil {
				// Best-effort: return introspection metadata when column definitions cannot be loaded.
				result.Result[rawObjectName] = *objectMetadata

				continue
			}

			for i := range columns {
				col := columns[i]
				objectMetadata.AddFieldMetadata(CustomFieldKey(col.ID), col.fieldMetadata())
			}
		}

		result.Result[rawObjectName] = *objectMetadata
	}

	return result, nil
}

func (c *Connector) fetchObjectMetadataViaIntrospection(
	ctx context.Context,
	objectName string,
) (*common.ObjectMetadata, error) {
	query, err := introspectionQueryForObject(objectName)
	if err != nil {
		return nil, err
	}

	res, err := c.postGraphQL(ctx, query)
	if err != nil {
		return nil, err
	}

	return c.parseSingleObjectMetadataResponse(ctx, objectName, nil, res)
}
