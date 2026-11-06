package core

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

const apiVersion = "v4_6_release/apis/3.0"

type Base struct {
	components.Connector

	common.RequireAuthenticatedClient
	common.RequireMetadata

	ErrorHandler common.ErrorHandler

	clientId string
}

func NewBase(params common.ConnectorParams) (*Base, error) {
	return components.Init(providers.ConnectWise, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Base, error) {
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle
	base.SetErrorHandler(errorHandler)

	return &Base{
		Connector:            *base,
		ErrorHandler:         errorHandler,
		ExpectedMetadataKeys: []string{"clientId"},
		clientId:             params.Metadata["clientId"],
	}, nil
}

func (b Base) ClientIdHeader() common.Header {
	return common.Header{
		Key:   "ClientId",
		Value: b.clientId,
	}
}

func (b Base) GetUrl(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, objectPath)
}

func (b Base) GetCommunicationTypesUrl() (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "company/communicationTypes")
}

func (b Base) GetSubscriptionUrl() (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "system/callbacks")
}
