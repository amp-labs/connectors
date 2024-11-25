package hubspot

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var (
	getRecordSupportedObjectsSet = datautils.NewStringSet(
		"company", "contact", "deal", "ticket", "line_item", "product",
	)

	errGerRecordNotSupportedForObject = errors.New("getRecord is not supproted for the object")
)

/*
   docs:
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/companies
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/contacts
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/deals
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/tickets
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/line_items
   https://developers.hubspot.com/beta-docs/reference/api/crm/objects/products
*/

// GetRecord returns a record from the object with the given ID and object name.
func (c *Connector) GetRecord(ctx context.Context, objectName string, recordId string) (*common.ReadResultRow, error) {
	if !getRecordSupportedObjectsSet.Has(objectName) {
		return nil, fmt.Errorf("%w %s", errGerRecordNotSupportedForObject, objectName)
	}

	pluralObjectName := naming.NewPluralString(objectName).String()
	relativePath := path.Join("/objects", pluralObjectName, recordId)

	resp, err := c.Client.Get(ctx, c.getURL(relativePath))
	if err != nil {
		return nil, err
	}

	record, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, fmt.Errorf("error parsing record: %w", err)
	}

	id, err := extractIdFromRecord(*record)
	if err != nil {
		// this should never happen unless the provider changes webhook message format
		return nil, err
	}

	return &common.ReadResultRow{
		Raw: *record,
		Id:  id,
	}, nil
}

var (
	errMissingId    = errors.New("missing id field in raw record")
	errTypeMismatch = errors.New("field is not a string")
)

//nolint:revive
func (c *Connector) GetRecordsWithIds(
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
) ([]common.ReadResultRow, error) {
	if !getRecordSupportedObjectsSet.Has(objectName) {
		return nil, fmt.Errorf("%w %s", errGerRecordNotSupportedForObject, objectName)
	}

	inputs := make([]map[string]any, len(ids))
	for i, id := range ids {
		inputs[i] = map[string]any{
			"properties": fields,
			"id":         id,
		}
	}

	pluralObjectName := naming.NewPluralString(objectName).String()
	relativePath := path.Join("/objects", pluralObjectName, "batch", "read")

	body := map[string]any{
		"inputs": inputs,
	}

	resp, err := c.Client.Post(ctx, c.getURL(relativePath), body)
	if err != nil {
		return nil, err
	}

	resBody, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := getRecords(resBody)
	if err != nil {
		return nil, err
	}

	if len(fields) != 0 {
		// If fields are specified, extract only those fields from the record.
		return getMarshalledData(records, fields)
	}

	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		id, err := extractIdFromRecord(record)
		if err != nil {
			// this should never happen unless the provider changes webhook message format
			return nil, err
		}

		data[i] = common.ReadResultRow{
			Raw: record,
			Id:  id,
		}
	}

	return data, nil
}

func extractIdFromRecord(record map[string]any) (string, error) {
	id, ok := record["id"]
	if !ok {
		return "", errMissingId
	}

	idStr, ok := id.(string)
	if !ok {
		return "", errTypeMismatch
	}

	return idStr, nil
}
