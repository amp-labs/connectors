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
		"company", "contact", "deal", "ticket", "line_item", "product", "user",
	)
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

var errMissingId = errors.New("missing id field in raw record")

//nolint:revive,funlen
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	singularObjName := naming.NewSingularString(objectName).String()
	if !getRecordSupportedObjectsSet.Has(singularObjName) {
		return nil, fmt.Errorf("%w %s", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	inputs := make([]map[string]any, len(ids))
	for i, id := range ids {
		inputs[i] = map[string]any{
			"id": id,
		}
	}

	pluralObjectName := naming.NewPluralString(objectName).String()

	url, err := c.getBatchRecordsURL(pluralObjectName, associations)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"inputs":     inputs,
		"properties": fields,
	}

	resp, err := c.Client.Post(ctx, url, body)
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

	return c.getDataMarshaller(ctx, objectName, associations)(records, fields)
}

func (c *Connector) getBatchRecordsURL(objectName string, associations []string) (string, error) {
	relativePath := strings.Join([]string{"/objects", objectName, "batch", "read"}, "/")

	if len(associations) > 0 {
		return c.getModuleURL(relativePath, "associations", strings.Join(associations, ","))
	}

	return c.getURLFromModule(relativePath)
}
