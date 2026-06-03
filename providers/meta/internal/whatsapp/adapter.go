package whatsapp

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

const (
	metadataKeyWhatsAppAccountID = "whatsappAccountId"
	metadataKeyPhoneNumberID     = "whatsappPhoneNumberId"
)

type Adapter struct {
	*components.Connector
	common.RequireMetadata
	components.Writer

	whatsappAccountId string
	phoneNumberId     string
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	adapter, err := components.Initialize(providers.Meta, params, constructor)
	if err != nil {
		return nil, err
	}

	adapter.whatsappAccountId = params.Metadata[metadataKeyWhatsAppAccountID]
	adapter.phoneNumberId = params.Metadata[metadataKeyPhoneNumberID]

	return adapter, nil
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{
				metadataKeyWhatsAppAccountID,
				metadataKeyPhoneNumberID,
			},
		},
	}

	registry := components.NewEmptyEndpointRegistry()
	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		registry,
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return adapter, nil
}
