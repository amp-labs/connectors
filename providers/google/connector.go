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

var (
	_ connectors.SubscribeConnector              = &Connector{}
	_ connectors.SubscriptionMaintainerConnector = &Connector{}
)

// GmailSubscribeRequest is the request payload for Gmail watch subscriptions.
type GmailSubscribeRequest = mail.WatchRequest

// GmailSubscribeResponse is the response payload from Gmail watch subscriptions.
type GmailSubscribeResponse = mail.WatchResponse

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

	if connector.Module() == providers.ModuleGoogleGmail {
		adapter, err := mail.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Mail = adapter
	}

	return connector, nil
}

func (c *Connector) ListObjectMetadata(
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

func (c *Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
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

func (c *Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
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

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
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

func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if c.Mail != nil {
		return c.Mail.Subscribe(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	if c.Mail != nil {
		return c.Mail.UpdateSubscription(ctx, params, previousResult)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) DeleteSubscription(
	ctx context.Context,
	previousResult common.SubscriptionResult,
) error {
	if c.Mail != nil {
		return c.Mail.DeleteSubscription(ctx, previousResult)
	}

	return common.ErrNotImplemented
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	if c.Mail != nil {
		return c.Mail.EmptySubscriptionParams()
	}

	return nil
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	if c.Mail != nil {
		return c.Mail.EmptySubscriptionResult()
	}

	return nil
}

func (c *Connector) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if c.Mail != nil {
		return c.Mail.VerifyWebhookMessage(ctx, request, params)
	}

	return false, common.ErrNotImplemented
}

func (c *Connector) GetRecordsByIds(ctx context.Context, // nolint: revive
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	if c.Mail != nil {
		return c.Mail.GetRecordsByIds(ctx, objectName, recordIds, fields, associations)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	if c.Mail != nil {
		return c.Mail.RunScheduledMaintenance(ctx, params, previousResult)
	}

	return nil, common.ErrNotImplemented
}

// Re-exports of Gmail history.list types so external callers can use them
// without importing the internal mail package.
type (
	HistoryListParams  = mail.HistoryListParams
	HistoryListResult  = mail.HistoryListResult
	HistoryRecord      = mail.HistoryRecord
	HistoryMessage     = mail.HistoryMessage
	HistoryMessageChange = mail.HistoryMessageChange
)

// HistoryList fetches Gmail mailbox changes since the given history checkpoint.
// Only valid when the connector is initialized for the Gmail module.
func (c *Connector) HistoryList(
	ctx context.Context, params HistoryListParams,
) (*HistoryListResult, error) {
	if c.Mail == nil {
		return nil, common.ErrNotImplemented
	}

	return c.Mail.HistoryList(ctx, params)
}

func (c *Connector) setUnitTestBaseURL(url string) {
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
