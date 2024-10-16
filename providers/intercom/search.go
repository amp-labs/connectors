package intercom

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Search is used when `Since` parameter is provided for the select few objects.
// Documentation below explains how search is done for "conversations" object. Others, use the same query language.
// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/searchcontacts
func (c *Connector) readViaSearch(
	ctx context.Context, config common.ReadParams,
) (*common.JSONHTTPResponse, *urlbuilder.URL, error, bool) {
	if config.Since.IsZero() {
		// Search is only relevant when we do incremental reading.
		return nil, nil, nil, false
	}

	if !incrementalSearchObjectPagination.Has(config.ObjectName) {
		return nil, nil, nil, false
	}

	url, err := constructURL(c.BaseURL, config.ObjectName, "search")
	if err != nil {
		return nil, nil, err, true
	}

	conversation, err := c.createSearchPayload(config)
	if err != nil {
		return nil, nil, err, true
	}

	rsp, err := c.Client.Post(ctx, url.String(), &conversation, apiVersionHeader)
	if err != nil {
		return nil, nil, err, true
	}

	return rsp, url, nil, true
}

func (c *Connector) createSearchPayload(params common.ReadParams) (*searchReqPayload, error) {
	url, err := urlbuilder.New(params.NextPage.String())
	if err != nil {
		return nil, err
	}

	// Unix time format is used.
	updatedAfter := strconv.FormatInt(params.Since.Unix(), 10)
	// We no longer request by GET, so query parameter must be moved to the POST payload.
	startingAfter, _ := url.GetFirstQueryParam("starting_after")

	conversation := searchReqPayload{
		Query: searchQuery{
			Operator: "AND",
			Value: []searchQueryValue{{
				Field:    "updated_at",
				Operator: ">",
				Value:    updatedAfter,
			}},
		},
		Pagination: searchPagination{
			PerPage:       incrementalSearchObjectPagination.Get(params.ObjectName),
			StartingAfter: startingAfter,
		},
	}

	return &conversation, nil
}

type searchReqPayload struct {
	Query      searchQuery      `json:"query"`
	Pagination searchPagination `json:"pagination"`
}

type searchQuery struct {
	Operator string             `json:"operator"`
	Value    []searchQueryValue `json:"value"`
}

type searchQueryValue struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type searchPagination struct {
	PerPage       int    `json:"per_page"`                 //nolint:tagliatelle
	StartingAfter string `json:"starting_after,omitempty"` //nolint:tagliatelle
}
