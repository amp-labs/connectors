package custom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) updateCustomFields(
	ctx context.Context, objectName string, groupName string,
	definitions []common.FieldDefinition, fields map[string]common.FieldUpsertResult,
) error {
	// There is no batch update, therefore for each definition we have to make a dedicated call.
	for _, definition := range definitions {
		if err := a.updateCustomField(ctx, objectName, groupName, definition, fields); err != nil {
			return err
		}
	}

	return nil
}

func (a *Adapter) updateCustomField(
	ctx context.Context, objectName string, groupName string,
	definition common.FieldDefinition, fields map[string]common.FieldUpsertResult,
) error {
	url, err := a.getPropertyUpdateURL(objectName, definition.FieldName)
	if err != nil {
		return err
	}

	payload, err := newPayload(groupName, definition)
	if err != nil {
		return err
	}

	response, err := a.makeRequestUpdate(ctx, url, payload)
	if err != nil {
		return err
	}

	response.populateField(fields, common.UpsertMetadataActionUpdate)

	return nil
}

func (a *Adapter) makeRequestUpdate(
	ctx context.Context, url *urlbuilder.URL, payload *Payload,
) (*Response, error) {
	responseData, err := a.Client.Patch(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[Response](responseData)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return response, nil
}
