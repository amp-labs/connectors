package stripe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
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

// Provider returns the provider this connector talks to. It is declared here rather than
// inherited from the embedded base because the subscribe registry registers a zero-value
// Connector as its webhook verifier, which has no base: the base's Provider has a value
// receiver, so the promoted call would dereference that nil pointer and panic.
func (c *Connector) Provider() providers.Provider {
	return providers.Stripe
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the zero-value verifier connector has no base to delegate to.
func (c *Connector) String() string {
	if c == nil || c.Base == nil {
		return c.Provider() + ".Connector"
	}

	return c.Base.String()
}
