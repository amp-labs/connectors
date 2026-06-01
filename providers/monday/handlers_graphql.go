package monday

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func buildGraphQLHTTPRequest(ctx context.Context, baseURL, query string) (*http.Request, error) {
	url, err := urlbuilder.New(baseURL, apiVersion)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]string{"query": query}

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

func buildReadQuery(params common.ReadParams) (string, error) {
	if params.ObjectName == mondayObjectItems {
		return buildItemsReadQuery(params)
	}

	return buildPaginatedReadQuery(params)
}

func buildItemsReadQuery(params common.ReadParams) (string, error) {
	boardID, err := boardIDFromReadParams(params)
	if err != nil {
		return "", err
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = defaultPageSize
	}

	cursor := ""
	if params.NextPage != "" {
		cursor = params.NextPage.String()
	}

	return getItemsQuery(boardID, limit, cursor, true), nil
}

func buildPaginatedReadQuery(params common.ReadParams) (string, error) {
	var page *int

	limit := 0

	if params.NextPage != "" {
		var pageNum int

		_, err := fmt.Sscanf(string(params.NextPage), "%d", &pageNum)
		if err != nil {
			return "", fmt.Errorf("invalid next page format: %w", err)
		}

		page = &pageNum
		limit = defaultPageSize
	}

	return getQueryForObject(params.ObjectName, page, &limit)
}
