package connectwise

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/amp-labs/connectors/providers/connectwise/internal/search"
	"github.com/amp-labs/connectors/providers/connectwise/internal/subscriber"
	"github.com/amp-labs/connectors/providers/connectwise/internal/webhook"
)

// Type Exports.
type (
	SubscriptionEvent          = webhook.Event
	CollapsedSubscriptionEvent = webhook.CollapsedSubscriptionEvent
	SubscriptionRequest        = subscriber.Request
	SubscriptionResponse       = subscriber.Result
	SubscriptionResource       = subscriber.SubscriptionResource
)

// SearchRawTypeFilterConditions is the type of raw filter search.
// This statement is an Export. Do not remove.
const SearchRawTypeFilterConditions = search.RawTypeFilterConditions

type (
	subscribeStrategy = subscriber.Strategy
	searchStrategy    = search.Strategy
)

var (
	_ connectors.ObjectMetadataConnector    = (*Connector)(nil)
	_ connectors.ReadConnector              = (*Connector)(nil)
	_ connectors.WriteConnector             = (*Connector)(nil)
	_ connectors.DeleteConnector            = (*Connector)(nil)
	_ connectors.BatchRecordReaderConnector = (*Connector)(nil)
	_ connectors.WebhookVerifierConnector   = (*Connector)(nil)
	_ connectors.SubscribeConnector         = (*Connector)(nil)
	_ connectors.SearchConnector            = (*Connector)(nil)
)

type Connector struct {
	*core.Base

	components.Reader
	components.Writer
	components.Deleter
	// TODO must use webhook.Verifier instead of webhook.NoopVerifier
	*webhook.NoopVerifier
	*subscribeStrategy
	*searchStrategy
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	base, err := core.NewBase(params)
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		Base: base,
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		common.ModuleRoot,
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  base.ErrorHandler,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		common.ModuleRoot,
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  base.ErrorHandler,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		common.ModuleRoot,
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  base.ErrorHandler,
		},
	)

	connector.NoopVerifier = webhook.NewVerifier(base)
	connector.subscribeStrategy = subscriber.NewStrategy(base)
	connector.searchStrategy = search.NewStrategy(base)

	return connector, nil
}

// Provider returns the provider this connector talks to. It is declared here rather than
// inherited from the embedded base because the subscribe registry registers a zero-value
// Connector as its webhook verifier, which has no base: the base's Provider has a value
// receiver, so the promoted call would dereference that nil pointer and panic.
func (c *Connector) Provider() providers.Provider {
	return providers.ConnectWise
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the zero-value verifier connector has no base to delegate to.
func (c *Connector) String() string {
	if c == nil || c.Base == nil {
		return c.Provider() + ".Connector"
	}

	return c.Base.String()
}
