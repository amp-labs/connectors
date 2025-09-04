package custom

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (a *Adapter) createCustomFields(
	ctx context.Context, objectName string, groupName string,
	definitions []common.FieldDefinition, fields map[string]common.FieldUpsertResult,
) ([]string, error) {
	url, err := a.getPropertyBatchCreateURL(objectName)
	if err != nil {
		return nil, err
	}

	payload, err := newBatchPayload(groupName, definitions)
	if err != nil {
		return nil, err
	}

	response, err := a.makeRequestCreate(ctx, url, payload)
	if err != nil {
		return nil, err
	}

	errorMessages := datautils.NewStringSet()
	fieldsForUpdate := make([]string, 0)

	// Duplicate field creation should be recorded and these fields will be used to attempt update.
	for _, errObject := range response.Errors {
		if errObject.Category == "OBJECT_ALREADY_EXISTS" {
			// This field should be updated instead.
			fieldsForUpdate = append(fieldsForUpdate, errObject.Context.Name...)
		} else {
			errorMessages.AddOne(errObject.Message)
		}
	}

	if len(errorMessages) != 0 {
		messages := errorMessages.List()
		sort.Strings(messages)

		return nil, fmt.Errorf("%w: %v", common.ErrBadRequest, strings.Join(messages, "; "))
	}

	response.Results.populateFields(fields, common.UpsertMetadataActionCreate)

	return fieldsForUpdate, nil
}

func (a *Adapter) makeRequestCreate(
	ctx context.Context, url *urlbuilder.URL, payload *BatchPayload,
) (*BatchResponse, error) {
	responseData, err := a.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[BatchResponse](responseData)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return response, nil
}
