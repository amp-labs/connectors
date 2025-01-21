package hubspot

import (
	"context"
	"encoding/json"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

// https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
//
//nolint:gochecknoglobals
var KnownObjectTypes = map[string]string{
	"0-2":   "companies",
	"0-1":   "contacts",
	"0-3":   "deals",
	"0-5":   "tickets",
	"0-421": "appointments",
	"0-48":  "calls",
	"0-18":  "communications",
	"0-410": "courses",
	"0-49":  "emails",
	"0-136": "leads",
	"0-8":   "line_items",
	"0-420": "listings",
	"0-54":  "marketing_events",
	"0-47":  "meetings",
	"0-46":  "notes",
	"0-116": "postal_mail",
	"0-7":   "products",
	"0-14":  "quotes",
	"0-162": "services",
	"0-69":  "subscriptions",
	"0-27":  "tasks",
	"0-115": "users",
}

func (c *Connector) GetSchema(ctx context.Context, objectNameOrTypeId string) (*common.StringMap, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	u := c.getSchemaURL(objectNameOrTypeId)

	resp, err := c.Client.Get(ctx, u)
	if err != nil {
		return nil, err
	}

	resBody, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	schema := make(common.StringMap)

	if err := json.Unmarshal(resBody.Source(), &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (c *Connector) getSchemaURL(objectNameOrTypeId string) string {
	return c.BaseURL + "/crm-object-schemas/v3/schemas/" + objectNameOrTypeId
}
