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
	ctx = logging.With(ctx, "provider", "netsuite", "step", "get_post_auth_info")
	log := logging.Logger(ctx)

	timezone, err := c.retrieveInstanceTimezone(ctx)

	isDefault := "false"

	if err != nil {
		log.Warn("failed to retrieve instance timezone, falling back to default",
			"error", err,
			"default", DefaultTimezone,
		)

		timezone, _ = time.LoadLocation(DefaultTimezone)
		isDefault = "true"
	}

	// Guard against nil timezone (which would .String() as "UTC" and silently break reads).
	if timezone == nil {
		log.Error("retrieveInstanceTimezone returned nil location without error — using default",
			"default", DefaultTimezone,
		)

		timezone, _ = time.LoadLocation(DefaultTimezone)
		isDefault = "true"
	}

	c.instanceTimezone = timezone

	catalogVars := map[string]string{
		"sessionTimezone":          timezone.String(),
		"sessionTimezoneIsDefault": isDefault,
	}

	log.Info("resolved instance timezone — this will be stored on the connection",
		"sessionTimezone", timezone.String(),
		"sessionTimezoneIsDefault", isDefault,
	)

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}

// retrieveInstanceTimezone queries the NetSuite instance to get its timezone
// using the SESSIONTIMEZONE function via SuiteQL.
func (c *Connector) retrieveInstanceTimezone(ctx context.Context) (*time.Location, error) {
	return RetrieveInstanceTimezone(ctx, c.ProviderInfo().BaseURL, c.JSONHTTPClient())
}

// RetrieveInstanceTimezone queries a NetSuite instance for its timezone using SuiteQL.
// Exported so the M2M connector can reuse this without duplicating the logic.
func RetrieveInstanceTimezone(
	ctx context.Context,
	baseURL string,
	client *common.JSONHTTPClient,
) (*time.Location, error) {
	log := logging.Logger(ctx)

	url, err := urlbuilder.New(baseURL, "/services/rest/query/v1/suiteql")
	if err != nil {
		return nil, fmt.Errorf("failed to build SuiteQL URL: %w", err)
	}

	log.Debug("querying SESSIONTIMEZONE via SuiteQL",
		"baseURL", baseURL,
		"fullURL", url.String(),
	)

	query := suiteQLQuery{
		Query: "SELECT SESSIONTIMEZONE AS timezone FROM DUAL",
	}

	resp, err := client.Post(ctx, url.String(), query, common.Header{
		Key:   "Prefer",
		Value: "transient",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute timezone query: %w", err)
	}

	location, err := parseTimezoneResponse(resp)
	if err != nil {
		log.Warn("failed to parse timezone response",
			"error", err,
			"statusCode", resp.Code,
		)

		return nil, err
	}

	log.Info("parsed timezone from SuiteQL response",
		"location", location.String(),
		"locationIsNil", location == nil,
		"statusCode", resp.Code,
	)

	return location, nil
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
		return nil, fmt.Errorf("%w: items array is empty", ErrNoTimezoneData)
	}

	item := tzResp.Items[0]
	timezone := item.getTimezone()

	if timezone == "" {
		return nil, fmt.Errorf("%w: timezone field=%q, expr1 field=%q",
			ErrNoTimezoneData, item.Timezone, item.Expr1)
	}

	// Parse the timezone string (e.g., "America/Los_Angeles") into a time.Location
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("time.LoadLocation(%q) failed: %w", timezone, err)
	}

	if location == nil {
		return nil, fmt.Errorf("time.LoadLocation(%q) returned nil without error", timezone)
	}

	return location, nil
}
