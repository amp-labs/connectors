package marketo

import (
	"context"
	"errors"
	"log/slog"
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

	// check if this is activities API, and if it's the first call
	// if yes make a paging token call with the Since value provided
	// And then add to url nextPageToken the token response from the
	// paging token resource.
	if params.ObjectName == "activities" && params.NextPage == "" && !params.Since.IsZero() {
		// make the call to pagingToken
		pagingTokenURL, err := c.getAPIURL(pagingURLSuffix)
		if err != nil {
			return nil, err
		}

		pagingTokenURL.WithQueryParam("sinceDatetime", params.Since.Format(time.RFC3339))

		resp, err := c.Client.Get(ctx, pagingTokenURL.String())
		if err != nil {
			return nil, err
		}

		pagingResponse, err := common.UnmarshalJSON[pagingTokenResponse](resp)
		if err != nil {
			return nil, err
		}

		slog.Info("made a pagingToken request", "response", *pagingResponse)

		url.WithQueryParam("nextPageToken", pagingResponse.NextPageToken)

		// Add list of required activityTypeIds.
		url.WithQueryParam("activityTypeIds", "1,2,3,6,7,8,9,10,11,12")
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() {
		switch c.Module.ID {
		case providers.ModuleMarketoAssets:
			fmtTime := params.Since.Format(time.RFC3339)
			url.WithQueryParam("earliestUpdatedAt", fmtTime)
			url.WithQueryParam("latestUpdatedAt", time.Now().Format(time.RFC3339))
		case providers.ModuleMarketoLeads:
			fallthrough
		case common.ModuleRoot:
			fallthrough
		default: // we currently don't support filtering in leads.
		}
	}

	return url, nil
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
		if err := addFilteringIDQueries(url, "1"); err != nil {
			return err
		}
	}

	// If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		if paginatesByIDs(params.ObjectName) {
			return addFilteringIDQueries(url, params.NextPage.String())
		}

		url.WithQueryParam("nextPageToken", params.NextPage.String())
	}

	return nil
}
