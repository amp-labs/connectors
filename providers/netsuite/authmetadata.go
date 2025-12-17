package netsuite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// Reference: https://timdietrich.me/blog/netsuite-suiteql-dates-times/

var (
	ErrFailedToGetTimezone = errors.New("failed to get timezone from NetSuite instance")
	ErrEmptyResponseBody   = errors.New("empty response body")
	ErrNoTimezoneData      = errors.New("no timezone data returned")
)

// DefaultTimezone is used as a fallback when we cannot retrieve the instance timezone.
// Pacific time is chosen because many NetSuite instances are US-based.
const DefaultTimezone = "America/Los_Angeles"

// GetPostAuthInfo retrieves the instance timezone using SuiteQL.
// This is called after authentication to discover instance-specific configuration.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	timezone, err := c.retrieveInstanceTimezone(ctx)

	isDefault := "false"

	if err != nil {
		// Fall back to Pacific time if we can't retrieve the timezone.
		// This is a reasonable default since Netsuite server times are
		// generally known to be in PT.
		logging.Logger(ctx).Warn("failed to retrieve NetSuite instance timezone, using default",
			"error", err.Error(),
			"defaultTimezone", DefaultTimezone,
		)

		timezone, _ = time.LoadLocation(DefaultTimezone)
		isDefault = "true"
	}

	c.instanceTimezone = timezone

	catalogVars := map[string]string{
		"sessionTimezone":          timezone.String(),
		"sessionTimezoneIsDefault": isDefault,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}

// retrieveInstanceTimezone queries the NetSuite instance to get its timezone
// using the SESSIONTIMEZONE function via SuiteQL.
func (c *Connector) retrieveInstanceTimezone(ctx context.Context) (*time.Location, error) {
	// Build the SuiteQL URL - we always use the SuiteQL endpoint for this query
	// regardless of which module is configured, since SuiteQL is the only way to
	// query SESSIONTIMEZONE.
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "/services/rest/query/v1/suiteql")
	if err != nil {
		return nil, fmt.Errorf("failed to build SuiteQL URL: %w", err)
	}

	// Query to get the session timezone
	query := suiteQLQuery{
		Query: "SELECT SESSIONTIMEZONE AS timezone FROM DUAL",
	}

	resp, err := c.JSONHTTPClient().Post(ctx, url.String(), query, common.Header{
		Key:   "Prefer",
		Value: "transient",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute timezone query: %w", err)
	}

	return parseTimezoneResponse(resp)
}

type suiteQLQuery struct {
	Query string `json:"q"`
}

type timezoneResponse struct {
	Items []timezoneItem `json:"items"`
}

type timezoneItem struct {
	// NetSuite inconsistently returns the timezone field - sometimes as "timezone"
	// (matching our alias) and sometimes as "expr1" (ignoring the alias).
	Timezone string `json:"timezone"`
	Expr1    string `json:"expr1"`
}

func (t timezoneItem) getTimezone() string {
	if t.Timezone != "" {
		return t.Timezone
	}

	return t.Expr1
}

func parseTimezoneResponse(resp *common.JSONHTTPResponse) (*time.Location, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, ErrEmptyResponseBody
	}

	tzResp, err := jsonquery.ParseNode[timezoneResponse](body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timezone response: %w", err)
	}

	if len(tzResp.Items) == 0 {
		return nil, ErrNoTimezoneData
	}

	timezone := tzResp.Items[0].getTimezone()
	if timezone == "" {
		return nil, ErrNoTimezoneData
	}

	// Parse the timezone string (e.g., "America/Los_Angeles") into a time.Location
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timezone %q: %w", timezone, err)
	}

	return location, nil
}
