package chilipiper

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	weekHours = 167
	start     = "start"
	end       = "end"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := conn.buildURL(config.ObjectName, readpageSize)
	if err != nil {
		return nil, err
	}

	if !config.Since.IsZero() && config.ObjectName == meetings {
		url.WithQueryParam(start, config.Since.Format(time.RFC3339))
		// Adds 7 days from the since given time, and use it as ed value.
		// timestamp above this is ot supported.
		// nolint: mnd
		url.WithQueryParam(end, config.Since.Add(weekHours*time.Hour).Format(time.RFC3339))
	}

	// Check if we're reading Next Page of Records.
	if len(config.NextPage) > 0 {
		url, err = urlbuilder.New(config.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	resp, err := conn.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		extractRecords(config.ObjectName),
		nextRecordsURL(url, config.ObjectName),
		common.GetMarshaledData,
		config.Fields,
	)
}
