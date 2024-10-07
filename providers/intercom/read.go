package intercom

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// TODO this has changed since the last time re-architecture was performed.

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
	if config.ObjectName == "conversations" && !config.Since.IsZero() {
		// Conversations with non-empty Since fallback to POST, searching for conversation by time.
		// https://developers.intercom.com/docs/references/rest-api/api.intercom.io/conversations/searchconversations
		url, err := constructURL(c.BaseURL, config.ObjectName, "search")
		if err != nil {
			return nil, nil, err
		}

		conversation, err := c.createSearchPayload(config.NextPage, config.Since)
		if err != nil {
			return nil, nil, err
		}

		rsp, err := c.Client.Post(ctx, url.String(), &conversation, apiVersionHeader)
		if err != nil {
			return nil, nil, err
		}

		return rsp, url, nil
	}

	// Default.
	// READ is done the usual way via GET, listing object.
	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, nil, err
	}

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

func (c *Connector) createSearchPayload(
	nextPageURL common.NextPageToken, since time.Time,
) (*searchReqPayload, error) {
	url, err := urlbuilder.New(nextPageURL.String())
	if err != nil {
		return nil, err
	}

	// Unix time format is used.
	updatedAfter := strconv.FormatInt(since.Unix(), 10)
	// We no longer request by GET, so query parameter must be moved to the POST payload.
	startingAfter, _ := url.GetFirstQueryParam("starting_after")

	conversation := searchReqPayload{
		Query: searchQuery{
			Operator: "AND",
			Value: []searchQueryValue{{
				Field:    "updated_at",
				Operator: ">=",
				Value:    updatedAfter,
			}},
		},
		Pagination: searchPagination{
			PerPage:       DefaultConversationsPageSize,
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
