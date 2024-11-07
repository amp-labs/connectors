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

	return &common.ReadResultRow{
		Raw: *record,
	}, nil
}
