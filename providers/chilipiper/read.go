package chilipiper

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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
		url.WithQueryParam("start", config.Since.Format(time.RFC3339))
		url.WithQueryParam("end", time.Now().Format(time.RFC3339))
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
		nextRecordsURL(url.String()),
		common.GetMarshaledData,
		config.Fields,
	)
}
