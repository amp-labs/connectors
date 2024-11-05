package hubspot

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/amp-labs/connectors/common"
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

//nolint:gochecknoglobals
var getRecordSupportedObjectsToPathMap = map[string]string{
	"company":   "companies",
	"contact":   "contacts",
	"deal":      "deals",
	"ticket":    "tickets",
	"line_item": "line_items",
	"product":   "products",
}

var ErrUnsupportedObject = errors.New("unsupported object")

func (c *Connector) GetRecord(ctx context.Context, objectName string, recordId string) (*common.ReadResultRow, error) {
	objectNameInPath, ok := getRecordSupportedObjectsToPathMap[objectName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedObject, objectName)
	}

	relativePath := path.Join("/objects", objectNameInPath, recordId)

	resp, err := c.Client.Get(ctx, c.getURL(relativePath))
	if err != nil {
		return nil, err
	}

	record, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, fmt.Errorf("error parsing record: %w", err)
	}

	return &common.ReadResultRow{
		Raw: *record,
	}, nil
}
