package mail

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Adapter struct {
	*components.Connector
	components.Reader
	components.Writer
	components.Deleter
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Init(providers.Google, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: adapter.interpretJSONError},
		HTML: &interpreter.DirectFaultyResponder{Callback: adapter.interpretHTMLError},
	}.Handle

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}

func (a *Adapter) getCoreObjectReadUrl(objectName string) (*urlbuilder.URL, error) {
	path, err := Schemas.FindURLPath(a.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, path)
}

func (a *Adapter) getReadUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectSentMessages:
		return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "/users/me/messages")
	case virtualObjectInboxMessages:
		return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "/users/me/messages")
	case objectNameDrafts, objectNameMessages:
		fallthrough // Core message objects have URLs in `schema.json`.
	default:
		return a.getCoreObjectReadUrl(objectName)
	}
}

func (a *Adapter) getWriteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectSentMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case virtualObjectInboxMessages:
		// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/send
		return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "/users/me/messages/send")
	case objectNameMessages:
		return nil, common.ErrOperationNotSupportedForObject
	default:
		// Write share the same URLs as read.
		return a.getCoreObjectReadUrl(objectName)
	}
}

func (a *Adapter) getDeleteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectSentMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case virtualObjectInboxMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case objectNameMessages:
		return nil, common.ErrOperationNotSupportedForObject
	default:
		// Delete share the same URLs as read.
		return a.getCoreObjectReadUrl(objectName)
	}
}

func (a *Adapter) getMessageUrl(messageID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "users/me/messages", messageID)
}
