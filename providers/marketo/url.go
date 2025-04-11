package marketo

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const restAPIPrefix = "rest" //nolint:gochecknoglobals

func (c *Connector) constructReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if err := constructURLQueries(url, params); err != nil {
		return nil, err
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() {
		switch c.Module() {
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

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
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
