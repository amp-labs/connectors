package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

const DateFormat = "2006-01-02"

type Connector struct {
	*core.Base
}

func NewConnector(params common.ConnectorParams) (conn *Connector, outErr error) {
	base, err := core.NewBase(params)
	if err != nil {
		return nil, err
	}

	return &Connector{Base: base}, nil
}
