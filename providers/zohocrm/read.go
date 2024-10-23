package zohocrm

import (
	"context"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// Read retrieves data based on the provided common.ReadParams configuration parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	// Just incase someone sends leads, Instead of Leads
	// All Objects are capitalized in their API names.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(obj)
	if err != nil {
		return nil, err
	}

	// Adds the fields requirement parameter.
	fields := strings.Join(config.Fields.List(), ",")
	url.WithQueryParam("fields", fields)

	// Add the `If-Modified-Since` header if provided.
	// All Objects(Modules in Zoho terms) supports this.
	modHeader := common.Header{
		Key:   "If-Modified-Since",
		Value: config.Since.Format(time.RFC3339),
	}

	res, err := c.Client.Get(ctx, url.String(), modHeader)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		common.GetRecordsUnderJSONPath("data"),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}
