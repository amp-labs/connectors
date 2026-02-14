package marketo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	url, err := c.constructSearchURL(ctx, params)
	if err != nil {
		// If this is the case, we return a zero records response.
		if errors.Is(err, ErrZeroRecords) {
			return &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			}, nil
		}

		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		nextRecordsURL(params.ObjectName, url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) constructSearchURL(ctx context.Context, params *common.SearchParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	for _, flt := range params.Filter.FieldFilters {
		url.WithQueryParam(flt.FieldName, fmt.Sprintf("%s", flt.Value))
	}

	if err := c.handleSearchActivitiesAPI(ctx, url, params); err != nil {
		return nil, err
	}

	if assetsObjects.Has(params.ObjectName) {
		url.WithQueryParam("maxReturn", strconv.Itoa(maxReturn))
	} else {
		url.WithQueryParam("BatchSize", strconv.Itoa(batchSize))
	}

	return url, nil
}

func (c *Connector) handleSearchActivitiesAPI(ctx context.Context, url *urlbuilder.URL, params *common.SearchParams,
) error {
	if params.ObjectName == activities { //nolint:nestif
		var activityIDs string

		// searching in activities, requires passing activityFields, with the other searching criteria.
		for _, flt := range params.Filter.FieldFilters {
			if flt.FieldName == "activityTypeIds" {
				ids, err := common.AssertType[string](flt.Value)
				if err != nil {
					return err
				}

				activityIDs = ids
			}

			// incase we're reading activities and the activityTypeIds is not supplied.
			// we error out.
			if activityIDs == "" {
				return ErrFilterInvalid
			}
		}

		url.WithQueryParam(activityTypeIDs, activityIDs)

		if err := c.addSearchActivityNextParam(ctx, url, params); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) addSearchActivityNextParam(ctx context.Context, url *urlbuilder.URL, params *common.SearchParams,
) error {
	if params.NextPage != "" {
		url.WithQueryParam(nextPageQuery, params.NextPage.String())

		return nil
	}

	// Manually setting the since timestamp to `1970-01-01` for retrieving
	// all lead activities in the instance.
	// Get initial paging token for first request
	token, err := c.getPagingToken(ctx, time.Unix(0, 0).UTC())
	if err != nil {
		return err
	}

	url.WithQueryParam(nextPageQuery, token)

	return nil
}
