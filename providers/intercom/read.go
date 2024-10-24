package intercom

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	rsp, url, err := c.performReadQuery(ctx, config)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

// There are 2 choices. Default usage of GET.
// Or we can do POST for conversations scoping by `Since` time.
func (c *Connector) performReadQuery(
	ctx context.Context, config common.ReadParams,
) (*common.JSONHTTPResponse, *urlbuilder.URL, error) {
	if rsp, url, err, searchable := c.readViaSearch(ctx, config); searchable {
		return rsp, url, err
	}

	// Default.
	// READ is done the usual way via GET, listing object.
	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, nil, err
	}

	url = enhanceReadWithQueryParams(url, config.ObjectName, config.Since)

	rsp, err := c.Client.Get(ctx, url.String(), apiVersionHeader)
	if err != nil {
		return nil, nil, err
	}

	return rsp, url, nil
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return constructURL(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	return url, nil
}

func enhanceReadWithQueryParams(url *urlbuilder.URL, objectName string, since time.Time) *urlbuilder.URL {
	if objectName == "activity_logs" && !since.IsZero() {
		url.WithQueryParam("created_at_after", strconv.FormatInt(since.Unix(), 10))
	}

	return url
}
