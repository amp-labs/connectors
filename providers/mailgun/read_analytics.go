package mailgun

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// logsRequestBody is the POST /v1/analytics/logs request payload. Start and End
// are omitted when the caller supplies no Since/Until so the API's own defaults
// apply (start: 1 day before now, end: now — per the OpenAPI spec).
type logsRequestBody struct {
	Start      string         `json:"start,omitempty"`
	End        string         `json:"end,omitempty"`
	Pagination logsPagination `json:"pagination"`
}

type logsPagination struct {
	Limit int    `json:"limit"`
	Sort  string `json:"sort,omitempty"`
	Token string `json:"token,omitempty"`
}

// logsTimeLayout matches Mailgun's documented log timestamp format
// (e.g. "Mon, 08 Jul 2024 00:00:00 -0000").
const logsTimeLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// buildLogsRequest constructs the POST request for the analytics/logs object.
// The pagination cursor (params.NextPage) is carried in the request body.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/logs
//
// Quirk: Mailgun rejects a start time older than the account's log-retention
// window with HTTP 400 ("start time ... is before log retention time ...").
// The retention period is plan-dependent (e.g. ~5 days on trial/sandbox
// accounts) and is not discoverable ahead of time. When the caller supplies no
// Since/Until we therefore omit start/end entirely and let the API's defaults
// (last 1 day) apply — always inside any plan's retention. Callers passing an
// explicit Since must keep it within their account's retention window.
func (c *Connector) buildLogsRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "analytics/logs")
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	payload := logsRequestBody{
		Pagination: logsPagination{
			Limit: mustPageSize(params),
			Sort:  "@timestamp:asc",
			Token: params.NextPage.String(),
		},
	}

	if !params.Since.IsZero() {
		payload.Start = params.Since.Format(logsTimeLayout)
	}

	if !params.Until.IsZero() {
		payload.End = params.Until.Format(logsTimeLayout)
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// tagsRequestBody is the POST /v1/analytics/tags request payload. The modern
// Tags API (tag group "Tags New") paginates via skip/limit carried in the body;
// limit default is 10, max 1000 per the OpenAPI spec.
type tagsRequestBody struct {
	Pagination tagsPagination `json:"pagination"`
}

type tagsPagination struct {
	Skip  int `json:"skip"`
	Limit int `json:"limit"`
}

// buildTagsRequest constructs the POST request for the analytics/tags object —
// the modern account-scoped Tags API replacing the deprecated GET
// /v3/{domain}/tags. The skip offset (params.NextPage) rides in the body.
func (c *Connector) buildTagsRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "analytics/tags")
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	payload := tagsRequestBody{
		Pagination: tagsPagination{
			Skip:  nextPageInt(params.NextPage),
			Limit: mustPageSize(params),
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
