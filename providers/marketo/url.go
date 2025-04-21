package marketo

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const ( //nolint:gochecknoglobals
	restAPIPrefix   = "rest"
	pagingURLSuffix = "activities/pagingtoken"
	activities      = "activities"
	sinceQuery      = "sinceDatetime"
	nextPageQuery   = "nextPageToken"
	activityTypeIDs = "activityTypeIds"
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

	if err := constructURLQueries(url, params); err != nil {
		return nil, err
	}

	// Check if this is an initial request to the Marketo Activities API.
	// For the first call (no NextPage token) with a Since timestamp,
	// fetch a paging token to ensure pagination starts from the correct time.
	// Then, append the token to the URL for subsequent pagination.
	if params.ObjectName == activities && !params.Since.IsZero() {
		if err := c.addActivityParams(ctx, url, params); err != nil {
			return nil, err
		}
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() {
		if c.Module.ID == providers.ModuleMarketoAssets {
			fmtTime := params.Since.Format(time.RFC3339)
			url.WithQueryParam("earliestUpdatedAt", fmtTime)
			url.WithQueryParam("latestUpdatedAt", time.Now().Format(time.RFC3339))
		}
	}

	return url, nil
}

func (c *Connector) addActivityParams(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	if params.NextPage != "" {
		url.WithQueryParam(nextPageQuery, params.NextPage.String())
	} else {
		pagingTokenURL, err := c.getAPIURL(pagingURLSuffix)
		if err != nil {
			return err
		}

		pagingTokenURL.WithQueryParam(sinceQuery, params.Since.Format(time.RFC3339))

		resp, err := c.Client.Get(ctx, pagingTokenURL.String())
		if err != nil {
			return err
		}

		pagingResponse, err := common.UnmarshalJSON[pagingTokenResponse](resp)
		if err != nil {
			return err
		}

		url.WithQueryParam(nextPageQuery, pagingResponse.NextPageToken)
	}

	// Add list of required activityTypeIds.
	url.WithQueryParam(activityTypeIDs, params.Filter)

	return nil
}

func (c *Connector) constructMetadataURL(objectName string) (*urlbuilder.URL, error) {
	path, ok := hasMetadataResource(objectName)
	if !ok {
		return c.getAPIURL(objectName)
	}

	return urlbuilder.New(c.BaseURL, path)
}

func addFilteringIDQueries(urlbuilder *urlbuilder.URL, startIdx string) error {
	ids := make([]string, batchSize)

	idx, err := strconv.Atoi(startIdx)
	if err != nil {
		return errors.Join(err, common.ErrNextPageInvalid)
	}

	for i := range ids {
		ids[i] = strconv.Itoa(idx + i)
	}

	queryIDs := strings.Join(ids, ",")
	urlbuilder.WithQueryParam("filterValues", queryIDs)
	urlbuilder.WithQueryParam("filterType", "id")

	return nil
}

func constructURLQueries(url *urlbuilder.URL, params common.ReadParams) error {
	if paginatesByIDs(params.ObjectName) && len(params.NextPage) == 0 {
		// For the initial API request, we start filtering from ID 1-300 to fetch the earliest records.
		// Subsequent requests will use the last received ID for pagination.
		if err := addFilteringIDQueries(url, "1"); err != nil {
			return err
		}
	}

	// If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		if paginatesByIDs(params.ObjectName) {
			return addFilteringIDQueries(url, params.NextPage.String())
		}

		url.WithQueryParam(nextPageQuery, params.NextPage.String())
	}

	return nil
}
