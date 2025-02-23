package helpscout

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := conn.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	resp, err := conn.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		getRecords(config.ObjectName),
		nextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
