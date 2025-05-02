package insightly

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/insightly/metadata"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize    = 500
	DefaultPageSizeStr = "500"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

/*
	Response format:

[

	{
	  "LEAD_ID": 78563840,
	  .....
	  "FIRST_NAME": "Katherine",
	  "LAST_NAME": "Nguyen",
	}

]

	Records array is situated usually at the root level of a response.
	The identifier key includes object name.
*/
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	fieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(fieldName),
		nextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func nextRecordsURL(url *urlbuilder.URL) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		skipStr, ok := url.GetFirstQueryParam("skip")
		if !ok {
			skipStr = "0"
		}

		skip, err := strconv.Atoi(skipStr)
		if err != nil {
			return "", err
		}

		newSkip := skip + DefaultPageSize
		url.WithQueryParam("skip", strconv.Itoa(newSkip))

		return url.String(), nil
	}
}
