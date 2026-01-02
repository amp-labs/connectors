package stripe

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
	"github.com/spyzhov/ajson"
)

// Read retrieves a list of items for a given object.
// Features:
//   - NextPage: Supported for those objects that Stripe paginates.
//   - Incremental Reading: The `Since` parameter is not supported.
//   - AssociatedObjects: This parameter allows fetching nested objects. You need to specify list of fields to expand.
//     For more details, refer to the Stripe documentation on expanding objects:
//     https://docs.stripe.com/api/expanding_objects
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(res,
		makeGetRecords(responseFieldName),
		makeNextRecordsURL(url),
		common.MakeMarshaledDataFunc(flattenMetadata),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	if !params.Since.IsZero() && incrementalObjects.Has(params.ObjectName) {
		url.WithQueryParam("created[gte]", strconv.FormatInt(params.Since.Unix(), 10))
	}

	// Deeply nested objects can be requested as part of a single API request.
	// Example: Query parameter "expand[]=data.customer" will expand the nested customer object.
	//
	// For more details, refer to the Stripe documentation:
	// https://docs.stripe.com/expand#how-it-works
	if len(params.AssociatedObjects) != 0 {
		expandTargets := make([]string, len(params.AssociatedObjects))
		for index, associate := range params.AssociatedObjects {
			expandTargets[index] = "data." + associate
		}

		url.WithQueryParamList("expand[]", expandTargets)
	}

	return url, nil
}

// makeGetRecords creates a NodeRecordsFunc that extracts records from the API response
// using the specified field name. The field name corresponds to the array field in
// Stripe's response that contains the list of records.
func makeGetRecords(responseFieldName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

var incrementalObjects = datautils.NewSet( // nolint:gochecknoglobals
	"accounts",
	"application_fees",
	"balance/history",
	"balance_transactions",
	"charges",
	"checkout/sessions",
	"coupons",
	"credit_notes",
	"customers",
	"disputes",
	"events",
	"file_links",
	"files",
	"forwarding/requests",
	"identity/verification_reports",
	"identity/verification_sessions",
	"invoiceitems",
	"invoices",
	"issuing/authorizations",
	"issuing/cardholders",
	"issuing/cards",
	"issuing/disputes",
	"issuing/transactions",
	"payment_intents",
	"payouts",
	"plans",
	"prices",
	"products",
	"promotion_codes",
	"radar/early_fraud_warnings",
	"radar/value_lists",
	"refunds",
	"reporting/report_runs",
	"reviews",
	"setup_intents",
	"shipping_rates",
	"subscription_schedules",
	"subscriptions",
	"tax_rates",
	"topups",
	"transfers",
	"treasury/financial_accounts",
)
