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
	return components.Init(providers.Meta, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{
				metadataKeyWhatsAppAccountID,
				metadataKeyPhoneNumberID,
			},
		},
		whatsappAccountId: params.Metadata[metadataKeyWhatsAppAccountID],
		phoneNumberId:     params.Metadata[metadataKeyPhoneNumberID],
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
