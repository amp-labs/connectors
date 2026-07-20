package stripe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/stripe/internal/core"
	"github.com/amp-labs/connectors/providers/stripe/internal/reader"
)

type (
	ReadParamsOpts = reader.ReadParamsOpts
	readerStrategy = reader.Strategy
)

type Connector struct {
	*core.Base

	// Dependent services.
	*readerStrategy
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	base, err := core.NewBase(params)
	if err != nil {
		return nil, err
	}

	return &Connector{
		Base:           base,
		readerStrategy: reader.NewStrategy(base),
	}, nil
}
