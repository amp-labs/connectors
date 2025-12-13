package cloudtalk

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/cloudtalk/metadata"
	"github.com/spyzhov/ajson"
)

type Response struct {
	ResponseData ResponseData `json:"responseData"`
}

type ResponseData struct {
	Data       []map[string]any `json:"data"`
	PageNumber int              `json:"pageNumber"`
	PageCount  int              `json:"pageCount"`
	Limit      int              `json:"limit"`
	ItemsCount int              `json:"itemsCount"`
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) getURL(objectName string) (string, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), objectName)
	if err != nil {
		return "", err
	}

	return c.ProviderInfo().BaseURL + path, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	node, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	data, err := extractReadData(node)
	if err != nil {
		return nil, err
	}

	nextPage, itemsCount, err := makeNextPageToken(node)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     itemsCount,
		Data:     data,
		NextPage: common.NextPageToken(nextPage),
		Done:     nextPage == "",
	}, nil
}

func extractReadData(node *ajson.Node) ([]common.ReadResultRow, error) {
	// Extract data
	dataNodes, err := jsonquery.New(node, "responseData").ArrayRequired("data")
	if err != nil {
		return nil, err
	}

	data := make([]common.ReadResultRow, len(dataNodes))
	for i, n := range dataNodes {
		m, err := jsonquery.Convertor.ObjectToMap(n)
		if err != nil {
			return nil, err
		}

		data[i] = common.ReadResultRow{
			Fields: m,
			Raw:    m,
		}
	}

	return data, nil
}

func makeNextPageToken(node *ajson.Node) (string, int64, error) {
	responseData := jsonquery.New(node, "responseData")

	itemsCount, err := responseData.IntegerWithDefault("itemsCount", 0)
	if err != nil {
		return "", 0, err
	}

	pageNumber, err := responseData.IntegerWithDefault("pageNumber", 0)
	if err != nil {
		return "", 0, err
	}

	pageCount, err := responseData.IntegerWithDefault("pageCount", 0)
	if err != nil {
		return "", 0, err
	}

	var nextPage string
	if pageNumber < pageCount {
		// CloudTalk uses 1-based indexing for pages usually, and returns current pageNumber.
		// Next page is simply current + 1
		nextPage = strconv.Itoa(int(pageNumber) + 1)
	}

	return nextPage, itemsCount, nil
}

func interpretError(res *http.Response, body []byte) error {
	return common.InterpretError(res, body)
}
