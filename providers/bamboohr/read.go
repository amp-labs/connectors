package bamboohr

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/bamboohr/metadata"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = "100"

	pageKey     = "page"
	pageSizeKey = "pageSize"

	cursorPageLimitKey = "page[limit]"
	cursorPageAfterKey = "page[after]"

	dateFormat                    = "2006-01-02"
	defaultWhosOutWindowDaysAhead = 14
)

// buildReadRequest assembles a GET for a supported read object.
//
// Incremental read is not supported: ReadParams.Since and ReadParams.Until are ignored.
// BambooHR list endpoints do not offer a shared updated-since (or equivalent) query
// parameter across the objects we read, and there is no consistent last-modified
// timestamp field to apply connector-side time filtering. The whos_out object
// requires start/end dates on the request; those come from the current date and a
// fixed forward window (see addDefaultWhosOutDateRange), not from Since/Until.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	addPaginationParams(endpointURL, params)

	return endpointURL, nil
}

func addPaginationParams(endpointURL *urlbuilder.URL, params common.ReadParams) {
	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)

	if params.ObjectName == objectWhosOut {
		addDefaultWhosOutDateRange(endpointURL)
	}

	if pageNumberReadObjects.Has(params.ObjectName) {
		endpointURL.WithQueryParam(pageKey, "1")

		if pageSizeReadObjects.Has(params.ObjectName) {
			endpointURL.WithQueryParam(pageSizeKey, pageSize)
		}

		return
	}

	if cursorReadObjects.Has(params.ObjectName) {
		endpointURL.WithQueryParam(cursorPageLimitKey, pageSize)
	}
}

func addDefaultWhosOutDateRange(endpointURL *urlbuilder.URL) {
	start := time.Now().UTC()

	endpointURL.WithQueryParam("start", start.Format(dateFormat))
	endpointURL.WithQueryParam("end", start.AddDate(0, 0, defaultWhosOutWindowDaysAhead).Format(dateFormat))
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	reqURL, err := urlbuilder.FromRawURL(req.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		nodeRecordsForRead(c, params.ObjectName),
		nextPageForObject(params.ObjectName, reqURL),
		readhelper.MakeMarshaledDataFuncWithId(nil, idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func nodeRecordsForRead(c *Connector, objectName string) common.NodeRecordsFunc {
	if objectName == objectMetaUsers {
		return keyedObjectRecordNodes()
	}

	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)

	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

// keyedObjectRecordNodes returns a NodeRecordsFunc for read responses whose body is a
// JSON object keyed by record id, not a wrapper with an array field. BambooHR uses this
// shape for meta_users (see test/read/meta-users.json). Each object value becomes one
// record node; iteration order is undefined.
func keyedObjectRecordNodes() common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		if !node.IsObject() {
			return nil, jsonquery.ErrNotObject
		}

		data, err := node.GetObject()
		if err != nil {
			return nil, err
		}

		records := make([]*ajson.Node, 0, len(data))
		for _, record := range data {
			records = append(records, record)
		}

		return records, nil
	}
}

func nextPageForObject(objectName string, reqURL *urlbuilder.URL) common.NextPageFunc {
	switch {
	case pageNumberReadObjects.Has(objectName):
		return makePageNumberNextURL(reqURL)
	case cursorReadObjects.Has(objectName):
		return makeCursorNextURL(reqURL)
	default:
		return noNextPage
	}
}

func makePageNumberNextURL(reqURL *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextHref, err := nextHrefFromLinks(node)
		if err != nil {
			return "", err
		}

		if nextHref != "" {
			return nextHref, nil
		}

		meta := jsonquery.New(node, "meta")

		page, err := meta.IntegerOptional("page")
		if err != nil {
			return "", err
		}

		totalPages, err := meta.IntegerOptional("totalPages")
		if err != nil {
			return "", err
		}

		if page == nil || totalPages == nil || *page >= *totalPages {
			return "", nil
		}

		reqURL.WithQueryParam(pageKey, strconv.FormatInt(*page+1, 10))

		return reqURL.String(), nil
	}
}

func makeCursorNextURL(reqURL *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextHref, err := nextHrefFromLinks(node)
		if err != nil {
			return "", err
		}

		if nextHref != "" {
			return nextHref, nil
		}

		nextCursor, err := jsonquery.New(node, "meta", "page").TextWithDefault("nextCursor", "")
		if err != nil {
			return "", err
		}

		if nextCursor == "" {
			return "", nil
		}

		reqURL.WithQueryParam(cursorPageAfterKey, nextCursor)

		return reqURL.String(), nil
	}
}

func nextHrefFromLinks(node *ajson.Node) (string, error) {
	next, err := jsonquery.New(node, "_links").ObjectOptional("next")
	if err != nil || next == nil {
		return "", err
	}

	return jsonquery.New(next).StrWithDefault("href", "")
}

func idFieldForObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectEmployees:
		return readhelper.NewIdField("employeeId")
	default:
		return readhelper.NewIdField("id")
	}
}

func noNextPage(_ *ajson.Node) (string, error) {
	return "", nil
}
