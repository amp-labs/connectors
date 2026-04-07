package microsoft

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

var (
	_ connectors.BatchRecordReaderConnector = (*Connector)(nil)
)

func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) ([]common.ReadResultRow, error) {
	//TODO implement me
	panic("implement me")
}
