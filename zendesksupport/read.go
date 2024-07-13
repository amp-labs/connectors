package zendesksupport

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	link, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, link.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		makeGetTotalSize(config.ObjectName),
		makeGetRecords(config.ObjectName),
		getNextRecordsURL,
		getMarshalledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return constructURL(config.NextPage.String())
	}

	// First page
	link, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	return link, nil
}
