package pipedrive

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
)

const data = "data"

// Read retrieves data based on the provided read parameters.
// https://developers.pipedrive.com/docs/api/v1
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	node, hasData := resp.Body()
	if !hasData {
		return &common.ReadResult{
			Rows: 0,
			Data: []common.ReadResultRow{},
			Done: true,
		}, nil
	}

	if c.moduleID == providers.PipedriveV2 && !v2SupportingUpdateSince.Has(config.ObjectName) {
		return manualIncrementalSync(node, data, config, "update_time", time.RFC3339, nextRecordsURL(url))
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(data),
		nextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// begin fetching objects at provided start date
	// Supporting objects are: Activities & Notes only.
	if !config.Since.IsZero() && c.moduleID == providers.PipedriveV1 {
		since := config.Since.UTC().Format(time.DateTime)
		url.WithQueryParam("start_date", since)
	}

	// v2 objects supports this param, except pipelines,products,stages
	// could use the manualFiltering as they all support sort by updateTime
	if !config.Since.IsZero() && c.moduleID == providers.PipedriveV2 {
		if v2SupportingUpdateSince.Has(config.ObjectName) {
			since := config.Since.UTC().Format(time.RFC3339)
			url.WithQueryParam("updated_since", since)
		} else {
			url.WithQueryParam("sort_by", "update_time")
			url.WithQueryParam("sort_direction", "desc")
		}
	}

	return url, nil
}

// Manual incremental synchronization implementation for Zoho Desk
//
// Zoho Desk lacks native incremental sync support. This function iterates through records
// and returns those created or updated after the specified timestamp.
func manualIncrementalSync(node *ajson.Node, recordsKey string, config common.ReadParams, //nolint:cyclop
	timestampKey string, timestampFormat string, nextPageFunc common.NextPageFunc,
) (*common.ReadResult, error) {
	records, nextPage, err := readhelper.FilterSortedRecords(node, recordsKey,
		config.Since, timestampKey, timestampFormat, nextPageFunc)
	if err != nil {
		return nil, err
	}

	rows, err := common.GetMarshaledData(records, config.Fields.List())
	if err != nil {
		return nil, err
	}

	var done bool
	if nextPage == "" {
		done = true
	}

	return &common.ReadResult{
		Rows:     int64(len(records)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     done,
	}, nil
}
