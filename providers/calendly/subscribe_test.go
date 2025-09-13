package calendly

import (
	"context"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/stretchr/testify/require"
)

func TestCalendlySubscriptionInterface(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	require.NoError(t, err)

	// Ensure connector implements SubscribeConnector interface
	var _ connectors.SubscribeConnector = conn
}

func TestCalendlySubscribe(t *testing.T) { //nolint:funlen
	t.Parallel()

	ctx := context.Background()
	
	// Mock server setup
	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/webhook_subscriptions"),
				},
				Then: mockserver.ResponseString(http.StatusCreated, `{
					"resource": {
						"uri": "https://api.calendly.com/webhook_subscriptions/AAAAAAAAAAAAAAAA",
						"callback_url": "https://example.com/webhook",
						"signing_key": "abcd1234",
						"events": ["invitee.created", "invitee.canceled"],
						"scope": "organization",
						"organization": "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB",
						"state": "active",
						"created_at": "2019-01-02T03:04:05.678123Z",
						"updated_at": "2019-01-02T03:04:05.678123Z"
					}
				}`),
			},
			{
				If: mockcond.And{
					mockcond.Method("GET"),
					mockcond.Path("/users/me"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"resource": {
						"uri": "https://api.calendly.com/users/AAAAAAAAAAAAAAAA",
						"current_organization": "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB"
					}
				}`),
			},
		},
		Default: mockserver.Response(http.StatusNotFound),
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	require.NoError(t, err)

	subscribeParams := common.SubscribeParams{
		Request: &CalendlySubscriptionRequest{
			URL:          "https://example.com/webhook",
			Events:       []CalendlySubscriptionEvent{CalendlyEventInviteeCreated, CalendlyEventInviteeCanceled},
			Scope:        "organization",
			Organization: "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB",
		},
	}

	result, err := conn.Subscribe(ctx, subscribeParams)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, common.SubscriptionStatusSuccess, result.Status)

	subscriptionResult, ok := result.Result.(*CalendlySubscriptionResult)
	require.True(t, ok)
	require.NotEmpty(t, subscriptionResult.URI)
}

func TestCalendlyDeleteSubscription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	// Mock server for DELETE requests
	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.MethodDELETE(),
				Then: mockserver.Response(http.StatusNoContent),
			},
			{
				If: mockcond.And{
					mockcond.Method("GET"),
					mockcond.Path("/users/me"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"resource": {
						"uri": "https://api.calendly.com/users/AAAAAAAAAAAAAAAA",
						"current_organization": "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB"
					}
				}`),
			},
		},
		Default: mockserver.Response(http.StatusNotFound),
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	require.NoError(t, err)

	previousResult := common.SubscriptionResult{
		Result: &CalendlySubscriptionResult{
			URI: "https://api.calendly.com/webhook_subscriptions/AAAAAAAAAAAAAAAA",
		},
	}

	err = conn.DeleteSubscription(ctx, previousResult)
	require.NoError(t, err)
}

func TestCalendlyUpdateSubscription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	// Mock server for DELETE and POST
	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.MethodDELETE(),
				Then: mockserver.Response(http.StatusNoContent),
			},
			{
				If: mockcond.MethodPOST(),
				Then: mockserver.ResponseString(http.StatusCreated, `{
					"resource": {
						"uri": "https://api.calendly.com/webhook_subscriptions/NEWSUBSCRIPTIONID",
						"callback_url": "https://example.com/updated-webhook",
						"signing_key": "updated1234",
						"events": ["invitee.created"],
						"scope": "organization",
						"organization": "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB",
						"state": "active",
						"created_at": "2019-01-02T03:04:05.678123Z",
						"updated_at": "2019-01-02T03:04:05.678123Z"
					}
				}`),
			},
			{
				If: mockcond.Method("GET"),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"resource": {
						"uri": "https://api.calendly.com/users/AAAAAAAAAAAAAAAA",
						"current_organization": "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB"
					}
				}`),
			},
		},
		Default: mockserver.Response(http.StatusNotFound),
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	require.NoError(t, err)

	updateParams := common.SubscribeParams{
		Request: &CalendlySubscriptionRequest{
			URL:          "https://example.com/updated-webhook",
			Events:       []CalendlySubscriptionEvent{CalendlyEventInviteeCreated},
			Scope:        "organization",
			Organization: "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB",
		},
	}

	previousResult := &common.SubscriptionResult{
		Result: &CalendlySubscriptionResult{
			URI: "https://api.calendly.com/webhook_subscriptions/AAAAAAAAAAAAAAAA",
		},
	}

	result, err := conn.UpdateSubscription(ctx, updateParams, previousResult)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, common.SubscriptionStatusSuccess, result.Status)

	subscriptionResult, ok := result.Result.(*CalendlySubscriptionResult)
	require.True(t, ok)
	require.Contains(t, subscriptionResult.URI, "NEWSUBSCRIPTIONID")
}

func TestCalendlySubscriptionValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	require.NoError(t, err)

	// Test invalid request type
	_, err = conn.Subscribe(ctx, common.SubscribeParams{
		Request: "invalid",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid request type")

	// Test invalid event type
	_, err = conn.Subscribe(ctx, common.SubscribeParams{
		Request: &CalendlySubscriptionRequest{
			URL:          "https://example.com/webhook",
			Events:       []CalendlySubscriptionEvent{"invalid.event"},
			Scope:        "organization",
			Organization: "https://api.calendly.com/organizations/BBBBBBBBBBBBBBBB",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid event type")
} 