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
	if !incrementalSearchObjectPagination.Has(config.ObjectName) {
		return nil, nil, nil, false
	}

	if isTimeSearch(config) && config.ObjectName != ticketsObjectName {
		// Search is only relevant when we do incremental reading.
		// Tickets is an exception. We can do full read only via POST.
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
	if isTimeSearch(params) && params.ObjectName == ticketsObjectName {
		// Perform full read for tickets using POST query.
		// This is a hack, query is designed to return all objects.
		return &searchReqPayload{
			Query: searchQuery{
				Operator: "OR",
				Value: []searchQueryValue{{
					Field:    "open",
					Operator: "=",
					Value:    "true",
				}, {
					Field:    "open",
					Operator: "=",
					Value:    "false",
				}},
			},
			Pagination: searchPagination{
				PerPage: incrementalSearchObjectPagination.Get(params.ObjectName),
			},
		}, nil
	}

	url, err := urlbuilder.New(params.NextPage.String())
	if err != nil {
		return nil, err
	}

	// We no longer request by GET, so query parameter must be moved to the POST payload.
	startingAfter, _ := url.GetFirstQueryParam("starting_after")

	conversation := searchReqPayload{
		Query: searchQuery{
			Operator: "AND",
			Value:    makeQueries(params),
		},
		Pagination: searchPagination{
			PerPage:       incrementalSearchObjectPagination.Get(params.ObjectName),
			StartingAfter: startingAfter,
		},
	}

	return &conversation, nil
}

// Unix time format is used for both Since & Until.
func makeQueries(params common.ReadParams) []searchQueryValue {
	queries := make([]searchQueryValue, 0)

	if !params.Since.IsZero() {
		updatedAfter := strconv.FormatInt(params.Since.Unix(), 10)
		queries = append(queries, searchQueryValue{
			Field:    "updated_at",
			Operator: ">",
			Value:    updatedAfter,
		})
	}

	if !params.Until.IsZero() {
		updatedBefore := strconv.FormatInt(params.Until.Unix(), 10)
		queries = append(queries, searchQueryValue{
			Field:    "updated_at",
			Operator: "<=",
			Value:    updatedBefore,
		})
	}

	return queries
}

func isTimeSearch(config common.ReadParams) bool {
	return config.Since.IsZero() && config.Until.IsZero()
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
