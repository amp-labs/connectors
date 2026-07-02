package jump

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	maxPageSize = 100
)

//go:embed graphql/*.graphql
var queryFiles embed.FS

type QueryParameters struct {
	First         int
	After         string
	UpdatedAfter  string
	UpdatedBefore string
	Fields        map[string]bool
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	pageSize := maxPageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	fields := make(map[string]bool, len(params.Fields))
	for field := range params.Fields {
		fields[field] = true
	}

	queryParams := QueryParameters{
		First:  pageSize,
		Fields: fields,
	}

	if params.NextPage != "" {
		queryParams.After = params.NextPage.String()
	}

	if !params.Since.IsZero() {
		queryParams.UpdatedAfter = datautils.Time.FormatRFC3339inUTC(params.Since)
	}

	if !params.Until.IsZero() {
		queryParams.UpdatedBefore = datautils.Time.FormatRFC3339inUTC(params.Until)
	}

	return c.buildGraphQLRequest(ctx, params.ObjectName, queryParams)
}

func (c *Connector) buildGraphQLRequest(
	ctx context.Context,
	objectName string,
	queryParams QueryParameters,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	query, err := graphql.Operation(queryFiles, "query", objectName, queryParams)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if _, ok := resp.Body(); ok {
		errorResp, err := common.UnmarshalJSON[ResponseError](resp)
		if err == nil && errorResp != nil {
			if checkErr := checkErrorInResponse(errorResp); checkErr != nil {
				return nil, checkErr
			}
		}
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("items", "data", params.ObjectName),
		makeNextRecordsURL(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

//	{
//		"data": {
//		  "meetings": {
//			"items": [
//			  ................
//			],
//			"pageInfo": {
//			  "hasNextPage": true,
//			  "hasPreviousPage": false,
//			  "startCursor": "g3QAAAABZAAGb2Zmc2V0YQA=",
//			  "endCursor": "g3QAAAABZAAGb2Zmc2V0YQk="
//			}
//		  }
//		}
//	  }
func makeNextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageInfo, err := jsonquery.New(node, "data", objectName).ObjectOptional("pageInfo")
		if err != nil {
			return "", err
		}

		if pageInfo == nil {
			return "", nil
		}

		hasNextPage, err := jsonquery.New(pageInfo).BoolOptional("hasNextPage")
		if err != nil {
			return "", err
		}

		if hasNextPage == nil || !*hasNextPage {
			return "", nil
		}

		return jsonquery.New(pageInfo).StrWithDefault("endCursor", "")
	}
}
