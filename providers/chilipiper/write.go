package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	writeURL, err := conn.buildWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		writeURL.AddPath(config.RecordId)
	}

	_, err = conn.Client.Post(ctx, writeURL.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
	}, nil
}
