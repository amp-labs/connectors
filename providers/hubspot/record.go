package hubspot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
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

var (
	errMissingId    = errors.New("missing id field in raw record")
	errTypeMismatch = errors.New("field is not a string")
)

//nolint:revive,funlen
func (c *Connector) GetRecordsWithIds(
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	singularObjName := naming.NewSingularString(objectName).String()
	if !getRecordSupportedObjectsSet.Has(singularObjName) {
		return nil, fmt.Errorf("%w %s", errGerRecordNotSupportedForObject, objectName)
	}

	inputs := make([]map[string]any, len(ids))
	for i, id := range ids {
		inputs[i] = map[string]any{
			"id": id,
		}
	}

	pluralObjectName := naming.NewPluralString(objectName).String()

	u, err := c.getBatchRecordsURL(pluralObjectName, associations)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"inputs":     inputs,
		"properties": fields,
	}

	resp, err := c.Client.Post(ctx, u, body)
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
		return c.getMarshalledData(ctx, objectName, associations)(records, fields)
	}

	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		id, err := extractIdFromRecord(record)
		if err != nil {
			// this should never happen unless the provider changes subscription event format
			return nil, err
		}

		data[i] = common.ReadResultRow{
			Raw: record,
			Id:  id,
		}
	}

	return data, nil
}

func (c *Connector) getBatchRecordsURL(objectName string, associations []string) (string, error) {
	relativePath := strings.Join([]string{"/objects", objectName, "batch", "read"}, "/")

	if len(associations) > 0 {
		return c.getURL(relativePath, "associations", strings.Join(associations, ","))
	} else {
		return c.getURL(relativePath)
	}
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
