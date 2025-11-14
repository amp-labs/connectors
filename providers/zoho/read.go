package zoho

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
)

// Read retrieves data based on the provided common.ReadParams configuration parameters.
// ref: https://www.zoho.com/crm/developer/docs/api/v6/get-records.html
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	switch c.moduleID { //nolint:exhaustive
	case providers.ModuleZohoDesk:
		return c.read(ctx, config, nil)
	case providers.ModuleZohoServiceDeskPlus:
		return c.servicedeskplusAdapter.Read(ctx, config)
	default:
		headers := constructHeaders(config)

		return c.read(ctx, config, headers)
	}
}

func (c *Connector) read(ctx context.Context, config common.ReadParams,
	headers []common.Header,
) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String(), headers...)
	if err != nil {
		return nil, err
	}

	node, hasData := res.Body()
	if !hasData {
		// There is a special case for most of the objects in zoho Desk
		// The API responds with an empty response body.
		// indicating there are no records to fetch.
		return &common.ReadResult{
			Rows: 0,
			Data: []common.ReadResultRow{},
			Done: true,
		}, nil
	}

	if c.moduleID == providers.ModuleZohoDesk {
		switch {
		case objectsSortableByCreatedTime.Has(config.ObjectName):
			return manualIncrementalSync(node, dataKey, config, createdTimeKey, timeLayout, getNextRecordsURLDesk(url))
		case objectsSortablebyModifiedTime.Has(config.ObjectName):
			return manualIncrementalSync(node, dataKey, config, modifiedTimeKey, timeLayout, getNextRecordsURLDesk(url))
		default:
			return common.ParseResult(res,
				common.ExtractRecordsFromPath(dataKey),
				getNextRecordsURLDesk(url),
				common.GetMarshaledData,
				config.Fields,
			)
		}
	}

	return common.ParseResult(res,
		extractRecordsFromPath(config.ObjectName),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func constructHeaders(config common.ReadParams) []common.Header {
	// Add the `If-Modified-Since` header if provided.
	// All Objects(or Modules in ZohoCRM terms) supports this.
	if !config.Since.IsZero() {
		return []common.Header{
			{
				Key:   "If-Modified-Since",
				Value: config.Since.Format(time.RFC3339),
			},
		}
	}

	return []common.Header{}
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
