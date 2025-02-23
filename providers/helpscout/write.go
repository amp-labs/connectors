package helpscout

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	write := conn.Client.Post

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedWriteObjects.Has(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := conn.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		url.AddPath(config.RecordId)

		write = conn.Client.Patch
	}

	_, err = write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	// The response is always an empty response body.
	return &common.WriteResult{
		Success: true,
	}, nil
}
