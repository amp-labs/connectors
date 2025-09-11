package google

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/google/internal/calendar"
	"github.com/amp-labs/connectors/providers/google/internal/contacts"
	"github.com/amp-labs/connectors/providers/google/internal/mail"
)

// Connector for Google provider.
// Each adapter corresponds to Google Module implementation.
// Only one adapter can be non-nil and will be delegated to on reading/writing operations.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	Calendar *calendar.Adapter
	Contacts *contacts.Adapter
	Mail     *mail.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Google, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if connector.Module() == providers.ModuleGoogleCalendar {
		adapter, err := calendar.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Calendar = adapter
	}

	if connector.Module() == providers.ModuleGoogleContacts {
		adapter, err := contacts.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Contacts = adapter
	}

	if connector.Module() == providers.ModuleGoogleMail {
		adapter, err := mail.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Mail = adapter
	}

	return connector, nil
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.Calendar != nil {
		return c.Calendar.ListObjectMetadata(ctx, objectNames)
	}

	if c.Contacts != nil {
		return c.Contacts.ListObjectMetadata(ctx, objectNames)
	}

	if c.Mail != nil {
		return c.Mail.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Read(ctx, params)
	}

	if c.Contacts != nil {
		return c.Contacts.Read(ctx, params)
	}

	if c.Mail != nil {
		return c.Mail.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Write(ctx, params)
	}

	if c.Contacts != nil {
		return c.Contacts.Write(ctx, params)
	}

	if c.Mail != nil {
		return c.Mail.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Delete(ctx, params)
	}

	if c.Contacts != nil {
		return c.Contacts.Delete(ctx, params)
	}

	if c.Mail != nil {
		return c.Mail.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) setUnitTestBaseURL(url string) {
	if c.Calendar != nil {
		c.Calendar.SetUnitTestBaseURL(url)
	}

	if c.Contacts != nil {
		c.Contacts.SetUnitTestBaseURL(url)
	}

	if c.Mail != nil {
		c.Mail.SetUnitTestBaseURL(url)
	}
}
