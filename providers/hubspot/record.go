package hubspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
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

//nolint:revive,funlen
func (c *Connector) GetRecordsByIds(ctx context.Context, params common.ReadByIdsParams) ([]common.ReadResultRow, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	singularObjName := naming.NewSingularString(params.ObjectName).String()
	if !getRecordSupportedObjectsSet.Has(singularObjName) {
		return nil, fmt.Errorf("%w %s", common.ErrGetRecordNotSupportedForObject, params.ObjectName)
	}

	inputs := make([]map[string]any, len(params.RecordIds))
	for i, id := range params.RecordIds {
		inputs[i] = map[string]any{
			"id": id,
		}
	}

	pluralObjectName := naming.NewPluralString(params.ObjectName).String()

	u, err := c.getBatchRecordsURL(pluralObjectName, params.AssociatedObjects)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"inputs":     inputs,
		"properties": params.Fields,
	}

	resp, err := c.Client.Post(ctx, u, body)
	if err != nil {
		return nil, err
	}

	resBody, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := core.GetRecords(resBody)
	if err != nil {
		return nil, err
	}

	return c.getDataMarshaller(ctx, params.ObjectName, params.AssociatedObjects)(records, params.Fields)
}

func (c *Connector) getBatchRecordsURL(objectName string, associations []string) (string, error) {
	relativePath := strings.Join([]string{"/objects", objectName, "batch", "read"}, "/")

	if len(associations) > 0 {
		return c.getURL(relativePath, "associations", strings.Join(associations, ","))
	} else {
		return c.getURL(relativePath)
	}
}
