package salesloft

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

//nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context, objectName string,
	recordIds []string, fields []string, associations []string,
) ([]common.ReadResultRow, error) {
	panic("unimplemented")
}
