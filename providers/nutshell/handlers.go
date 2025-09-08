package nutshell

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/nutshell/internal/metadata"
	"github.com/spyzhov/ajson"
)

// The API accepts large page numbers without an error.
// The actual maximum limit is 10,000; anything higher will return an error.
const defaultPageSize = "10000"

// Events: https://developers.nutshell.com/reference/get_events
const objectNameEvents = "events"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// The Events endpoint uses parameter names that differ from other endpoints.
	if params.ObjectName == objectNameEvents {
		url.WithQueryParam("limit", defaultPageSize)
	} else {
		url.WithQueryParam("page[limit]", defaultPageSize)
	}

	// Time queries need a range.
	// Giving an Until without a Since is ignored.
	// Passing only Since will make Until default to now.
	// https://developers.nutshell.com/docs/filters#data-filters
	if incrementalReadByCreatedTimeQP.Has(params.ObjectName) && !params.Since.IsZero() {
		timeUntil := params.Until
		if timeUntil.IsZero() {
			timeUntil = time.Now()
		}

		url.WithQueryParam("filter[createdTime]", fmt.Sprintf("%v %v",
			datautils.Time.FormatRFC3339inUTC(params.Since),
			datautils.Time.FormatRFC3339inUTC(timeUntil),
		))
	}

	// The Events endpoint uses parameter names that differ from other endpoints.
	if params.ObjectName == objectNameEvents {
		if !params.Since.IsZero() {
			url.WithQueryParam("since_time", datautils.Time.Unix(params.Since))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("max_time", datautils.Time.Unix(params.Until))
		}
	}

	return url, nil
}

var incrementalReadByCreatedTimeQP = datautils.NewSet( // nolint:gochecknoglobals
	"accounts",
	"activities",
	"contacts",
	"leads",
)

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		// This response structure relates to `events` object.
		nextToken, err := jsonquery.New(node, "meta").StrWithDefault("next", "")
		if err != nil {
			return "", err
		}

		if nextToken != "" {
			return nextToken, nil
		}

		page, exists := url.GetFirstQueryParam("page[page]")
		if !exists {
			// First page count begins from zero. Current page is zeroth page.
			page = "0"
		}

		pageNum, err := strconv.Atoi(page)
		if err != nil {
			return "", err
		}

		// Advance pagination counter.
		pageNum += 1

		url.WithQueryParam("page[page]", strconv.Itoa(pageNum))

		return url.String(), nil
	}
}
