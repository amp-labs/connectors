package kit

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/kit/metadata"
)

// nolint:gochecknoglobals
// list of object names which support custom fields, currently only subscribers.
// custom fields belong to subscribers rather than forms, sequences, or tags.
// The GET API requests do not attach custom fields directly to forms, sequences, or tags.
var objectsWithCustomFields = datautils.NewStringSet(objectNameSubscribers)

// ListObjectMetadata creates metadata of object via reading objects using Kit API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult, err := metadata.Schemas.Select(c.Module.ID, objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		if !objectsWithCustomFields.Has(objectName) {
			continue
		}
		// Get a reference to the metadata in the map so changes are persisted
		objectMetadata, ok := metadataResult.Result[objectName]
		if !ok {
			// Object not found in result, skip it
			continue
		}

		fields, err := c.requestCustomFields(ctx)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}
		// initialize FieldsMap if it's nil.
		// Using deprecated FieldsMap instead of Fields to maintain consistency with
		// the metadata format returned by metadata Schemas Select.
		if objectMetadata.FieldsMap == nil { // nolint:staticcheck
			objectMetadata.FieldsMap = make(map[string]string) // nolint:staticcheck
		}

		for _, field := range fields {
			objectMetadata.FieldsMap[field.Key] = field.Label // nolint:staticcheck
		}

		metadataResult.Result[objectName] = objectMetadata
	}

	return metadataResult, nil
}

// requestCustomFields makes an API call to get custom fields for subscribers.
func (c *Connector) requestCustomFields(ctx context.Context) (map[string]customFieldDefinition, error) {
	url, err := c.getApiURL(objectNameCustomFields)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[customFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if fieldsResponse == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	fields := make(map[string]customFieldDefinition)
	for _, field := range fieldsResponse.CustomFields {
		fields[field.Label] = field
	}

	return fields, nil
}

type customFieldsResponse struct {
	CustomFields []customFieldDefinition `json:"custom_fields"`
}

type customFieldDefinition struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Key   string `json:"key"`
	Label string `json:"label"`
}
