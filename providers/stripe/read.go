package stripe

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
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

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
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
		common.GetOptionalRecordsUnderJSONPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	// Deeply nested objects can be requested as part of a single API request.
	// Example: Query parameter "expand[]=data.customer" will expand the nested customer object.
	//
	// For more details, refer to the Stripe documentation:
	// https://docs.stripe.com/expand#how-it-works
	if len(config.AssociatedObjects) != 0 {
		expandTargets := make([]string, len(config.AssociatedObjects))
		for index, associate := range config.AssociatedObjects {
			expandTargets[index] = "data." + associate
		}

		url.WithQueryParamList("expand[]", expandTargets)
	}

	return url, nil
}
