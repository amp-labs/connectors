package marketo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const ( //nolint:gochecknoglobals
	// API path components.
	restAPIPrefix   = "rest"
	activities      = "activities"
	leads           = "leads"
	pagingURLSuffix = "activities/pagingtoken"

	// URL parameter keys.
	idFilter             = "id"
	sinceQuery           = "sinceDatetime"
	nextPageQuery        = "nextPageToken"
	activityTypeIDs      = "activityTypeIds"
	filterValuesQuery    = "filterValues"
	filterTypeQuery      = "filterType"
	earliestUpdatedQuery = "earliestUpdatedAt"
	latestUpdatedAtQuery = "latestUpdatedAt"

	newLeadActivityType = "12"
	startingIDIdx       = "1"
)

type pagingTokenResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Success       bool   `json:"success"`
}

func (c *Connector) constructReadURL(ctx context.Context, params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if err := c.constructURLQueries(ctx, url, params); err != nil {
		return nil, err
	}

	if err := c.handleActivitiesAPI(ctx, url, params); err != nil {
		return nil, err
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() && c.Module.ID == providers.ModuleMarketoAssets {
		fmtTime := params.Since.Format(time.RFC3339)
		url.WithQueryParam(earliestUpdatedQuery, fmtTime)
		url.WithQueryParam(latestUpdatedAtQuery, time.Now().Format(time.RFC3339))
	}

	return url, nil
}

func (c *Connector) constructURLQueries(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	// We don't handle this scenario here, this is handled in the handleLeadsAPI
	// function, this indicates first call for the Incremental Lead Read.
	if params.ObjectName == leads && !params.Since.IsZero() && params.NextPage == "" {
		return c.handleLeadsAPI(ctx, url, params)
	}

	if paginatesByIDs(params.ObjectName) {
		switch len(params.NextPage) {
		case 0:
			// For the initial API request, we start filtering from ID 1-300 to fetch the earliest records.
			// Subsequent requests will use the last received ID for pagination.
			if err := addFilteringIDQueries(url, startingIDIdx); err != nil {
				return err
			}
		default:
			// For reading next-page requests, we add 300 filtering ids.
			// by ading +1 to the last record id.
			return addFilteringIDQueries(url, params.NextPage.String())
		}

		url.WithQueryParam(nextPageQuery, params.NextPage.String())
	}

	return nil
}

func (c *Connector) handleActivitiesAPI(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	// Check if this is an initial request to the Marketo Activities API.
	// For the first call (no NextPage token) with a Since timestamp,
	// fetch a paging token to ensure pagination starts from the correct time.
	// Then, append the token to the URL for subsequent pagination.
	if params.ObjectName == activities && !params.Since.IsZero() {
		if params.Filter == "" {
			return ErrFilterInvalid
		}

		url.WithQueryParam(activityTypeIDs, params.Filter)

		if err := c.addActivityNextParam(ctx, url, params); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) handleLeadsAPI(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	start, err := c.generateLeadStartID(ctx, params)
	if err != nil {
		return err
	}

	if err := addFilteringIDQueries(url, start); err != nil {
		return err
	}

	return nil
}

func (c *Connector) addActivityNextParam(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	if params.NextPage != "" {
		url.WithQueryParam(nextPageQuery, params.NextPage.String())

		return nil
	}

	// Get initial paging token for first request
	token, err := c.getPagingToken(ctx, params.Since)
	if err != nil {
		return err
	}

	url.WithQueryParam(nextPageQuery, token)

	return nil
}

func (c *Connector) generateLeadStartID(ctx context.Context, params common.ReadParams) (string, error) {
	token, err := c.getPagingToken(ctx, params.Since)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve paging token when reading leads: %w", err)
	}

	resp, err := c.getLeadActivities(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve activities when reading leads: %w", err)
	}

	if len(resp.Result) == 0 {
		return "", ErrZeroRecords
	}

	startIdx := strconv.Itoa(resp.Result[0].LeadID)

	return startIdx, nil
}

func (c *Connector) getLeadActivities(ctx context.Context, token string) (*readResponse, error) {
	url, err := c.getAPIURL(activities)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(activityTypeIDs, newLeadActivityType)
	url.WithQueryParam(nextPageQuery, token)

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	rsp, err := common.UnmarshalJSON[readResponse](res)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (c *Connector) getPagingToken(ctx context.Context, since time.Time) (string, error) {
	pagingTokenURL, err := c.getAPIURL(pagingURLSuffix)
	if err != nil {
		return "", err
	}

	pagingTokenURL.WithQueryParam(sinceQuery, since.Format(time.RFC3339))

	resp, err := c.Client.Get(ctx, pagingTokenURL.String())
	if err != nil {
		return "", err
	}

	pagingResponse, err := common.UnmarshalJSON[pagingTokenResponse](resp)
	if err != nil {
		return "", err
	}

	return pagingResponse.NextPageToken, nil
}

func (c *Connector) constructMetadataURL(objectName string) (*urlbuilder.URL, error) {
	path, ok := hasMetadataResource(objectName)
	if !ok {
		return c.getAPIURL(objectName)
	}

	return urlbuilder.New(c.BaseURL, path)
}

func addFilteringIDQueries(urlbuilder *urlbuilder.URL, startIdx string) error {
	start, err := strconv.Atoi(startIdx)
	if err != nil {
		return errors.Join(err, common.ErrNextPageInvalid)
	}

	ids := make([]string, batchSize)
	for i := range ids {
		ids[i] = strconv.Itoa(start + i)
	}

	urlbuilder.WithQueryParam(filterValuesQuery, strings.Join(ids, ","))
	urlbuilder.WithQueryParam(filterTypeQuery, idFilter)

	return nil
}
