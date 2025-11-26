package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	return nil, errors.New("attio doesn't support batch read by IDs") //nolint:err113
}
